package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/darmiel/golang-ical"
	"github.com/go-redis/redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/imroc/req/v3"
	"github.com/ralf-life/engine/engine"
	"github.com/ralf-life/engine/model"
	"gopkg.in/yaml.v3"
	"strings"
	"time"
)

// client requests the source URLs
var client = req.C().SetTimeout(10 * time.Second)

func (d *DemoServer) getSourceWithRequest(url string, cache time.Duration) (string, error) {
	// request source
	resp, err := client.R().Get(url)
	if err != nil {
		return "", err
	}
	val, err := resp.ToString()
	if err != nil {
		return "", err
	}
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
	debugMessages := make([]string, len(cp.Debugs))
	for i, v := range cp.Debugs {
		debugMessages[i] = fmt.Sprintf("%+v", v)
	}
	ctx.Append("X-Debug-Messages", debugMessages...)

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
