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
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// client requests the source URLs
var client = &http.Client{
	Timeout: 20 * time.Second,
}

const MaxContentLength = 8 * 1000 * 1000

var ErrExceededContentLength = errors.New("content-length exceeded limit of 8 MB")

func (d *DemoServer) getSourceWithRequest(url string, cache time.Duration) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.ContentLength > MaxContentLength {
		return "", ErrExceededContentLength
	}
	valBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	val := string(valBytes)

	err = d.red.SetEx(context.TODO(), "source::"+url, val, cache).Err()
	fmt.Println("[" + url + "] from request")
	return val, err
}

func (d *DemoServer) getSource(url string, cache time.Duration) (string, error) {
	body, err := d.red.Get(context.TODO(), "source::"+url).Result()
	if err != nil && err == redis.Nil {
		return d.getSourceWithRequest(url, cache)
	}
	fmt.Println("[" + url + "] from cache")
	return body, err
}

func (d *DemoServer) routeProcessDo(content []byte, ctx *fiber.Ctx) error {
	// try to parse body
	var profile model.Profile
	dec := yaml.NewDecoder(bytes.NewReader(content))
	dec.KnownFields(true)
	if err := dec.Decode(&profile); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid RALF-SPEC ("+err.Error()+")")
	}

	// validate profile
	if strings.TrimSpace(profile.Source) == "" {
		return fiber.NewError(fiber.StatusBadRequest, "`source` required")
	}

	// require a cache duration of at least 120s
	cd := time.Duration(profile.CacheDuration)
	if cd.Minutes() < 2.0 {
		cd = 2 * time.Minute
	}

	body, err := d.getSource(profile.Source, cd)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "cannot request source ("+err.Error()+")")
	}

	// parse calendar
	cal, err := ics.ParseCalendar(strings.NewReader(body))
	if err != nil {
		return fiber.NewError(fiber.StatusExpectationFailed, "failed to parse source calendar ("+err.Error()+")")
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
