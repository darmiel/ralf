package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/darmiel/golang-ical"
	"github.com/gofiber/fiber/v2"
	"github.com/ralf-life/engine/pkg/engine"
	"github.com/ralf-life/engine/pkg/model"
	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// client requests the source URLs
var client = &http.Client{
	Timeout: 20 * time.Second,
}

func (d *DemoServer) getSourceWithRequest(
	ctx context.Context,
	source model.Source,
	cache time.Duration,
	cacheKey string,
) (*ics.Calendar, error) {
	cal, err := source.Run()
	if err != nil {
		return nil, err
	}
	if err = d.red.SetEx(ctx, cacheKey, cal.Serialize(), cache).Err(); err != nil {
		return nil, err
	}
	return cal, nil
}

func (d *DemoServer) getSource(ctx context.Context, source model.Source, cache time.Duration) (*ics.Calendar, error) {
	cacheKey, err := source.CacheKey()
	if err != nil {
		return nil, err
	}
	body, err := d.red.Get(ctx, cacheKey).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			return nil, err
		}
		return d.getSourceWithRequest(ctx, source, cache, cacheKey)
	}
	return ics.ParseCalendar(strings.NewReader(body))
}

func (d *DemoServer) routeProcessDo(content []byte, ctx *fiber.Ctx) error {
	// try to parse body
	var profile model.Profile
	dec := yaml.NewDecoder(bytes.NewReader(content))
	dec.KnownFields(true)
	if err := dec.Decode(&profile); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid RALF-SPEC ("+err.Error()+")")
	}

	if len(profile.Source) != 1 {
		return fiber.NewError(fiber.StatusBadRequest, "only one source is allowed")
	}
	source := profile.Source[0]

	if err := source.Validate(); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid source ("+err.Error()+")")
	}

	// require a cache duration of at least 120s
	cd := time.Duration(profile.CacheDuration)
	if cd.Minutes() < 2.0 {
		cd = 2 * time.Minute
	}

	cal, err := d.getSource(ctx.Context(), source, cd)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "cannot request source ("+err.Error()+")")
	}

	// create context and run flow
	cp := &engine.ContextFlow{
		Profile:     &profile,
		Context:     make(map[string]interface{}),
		EnableDebug: true,
		Verbose:     true,
	}
	if err = engine.ModifyCalendar(cp, profile.Flows, cal); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to run flow ("+err.Error()+")")
	}

	// append debug messages as header
	ctx.Append("X-Debug-Message-Count", strconv.Itoa(len(cp.Debugs)))
	for i, v := range cp.Debugs {
		ctx.Append(fmt.Sprintf("X-Debug-Message-%d", i+1), fmt.Sprintf("%+v", v))
	}

	// append content-type and return calendar
	ctx.Set("Content-Type", "text/calendar")
	return ctx.Status(201).SendString(cal.Serialize())
}

func (d *DemoServer) routeProcessGet(ctx *fiber.Ctx) error {
	q := ctx.Query("tpl")
	if q == "" {
		return fiber.NewError(fiber.StatusBadRequest, "`tpl` (base64) parameter missing.")
	}
	content, err := base64.StdEncoding.DecodeString(q)
	if err != nil {
		return fiber.NewError(fiber.StatusExpectationFailed, "invalid base64 ("+err.Error()+")")
	}
	return d.routeProcessDo(content, ctx)
}

func (d *DemoServer) routeProcessPost(ctx *fiber.Ctx) error {
	return d.routeProcessDo(ctx.Body(), ctx)
}
