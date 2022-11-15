package server

import (
	"bytes"
	"context"
	"fmt"
	ics "github.com/arran4/golang-ical"
	"github.com/go-redis/redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/imroc/req/v3"
	"github.com/ralf-life/engine/actions"
	"github.com/ralf-life/engine/engine"
	"github.com/ralf-life/engine/model"
	"gopkg.in/yaml.v3"
	"strings"
	"time"
)

var client = req.C().SetTimeout(2 * time.Second)

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

func (d *DemoServer) routeProcess(ctx *fiber.Ctx) error {

	// try to parse body
	var profile model.Profile
	dec := yaml.NewDecoder(bytes.NewReader(ctx.Body()))
	dec.KnownFields(true)
	if err := dec.Decode(&profile); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid RALF-SPEC ("+err.Error()+")")
	}

	// require a cache duration of 60s
	cd := time.Duration(profile.CacheDuration)
	if cd.Minutes() < 1.0 {
		cd = time.Minute
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
	cal.SetMethod(ics.MethodRequest)

	cp := engine.ContextFlow{
		Profile: &profile,
		Context: make(map[string]interface{}),
	}

	// get components from calendar (events) and copy to slice for later modifications
	cc := cal.Components[:]

	// start from behind so we can remove from slice
	for i := len(cc) - 1; i >= 0; i-- {
		event, ok := cc[i].(*ics.VEvent)
		if !ok {
			continue
		}
		var fact actions.ActionMessage
		fact, err = cp.RunAllFlows(event, profile.Flows)
		if err != nil {
			if err == engine.ErrExited {
				fmt.Println("--> flows exited because of a return statement.")
			} else {
				return fiber.NewError(fiber.StatusInternalServerError, "failed to run flow ("+err.Error()+")")
			}
		}
		switch fact.(type) {
		case actions.FilterOutMessage:
			cc = append(cc[:i], cc[i+1:]...) // remove event from components
		}
	}

	cal.Components = cc
	return ctx.Status(201).SendString(cal.Serialize())
}
