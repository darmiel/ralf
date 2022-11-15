package engine

import (
	"fmt"
	ics "github.com/arran4/golang-ical"
	"strconv"
	"strings"
	"time"
)

type ContextEnv struct {
	Event *ctxEvent
	Date  *ctxTime
	Start *ctxTime
	End   *ctxTime
}

func (e *ContextEnv) AORB(val bool, a, b string) string {
	if val {
		return a
	}
	return b
}

func (e *ContextEnv) String(obj any) string {
	return fmt.Sprintf("%+v", obj)
}

type (
	ctxTime struct {
		*time.Time
	}
	ctxEvent struct {
		*ics.VEvent
	}
)

func (c *ctxTime) IsMonday() bool {
	return c.Weekday() == time.Monday
}

func (c *ctxTime) IsTuesday() bool {
	return c.Weekday() == time.Tuesday
}

func (c *ctxTime) IsWednesday() bool {
	return c.Weekday() == time.Wednesday
}

func (c *ctxTime) IsThursday() bool {
	return c.Weekday() == time.Thursday
}

func (c *ctxTime) IsFriday() bool {
	return c.Weekday() == time.Friday
}

func (c *ctxTime) IsSaturday() bool {
	return c.Weekday() == time.Saturday
}

func parseTime(time string) (hour, min int, ok bool) {
	var hourS, minS string
	if strings.Contains(time, ":") {
		spl := strings.Split(time, ":")
		hourS = spl[0]
		if len(spl) > 1 {
			minS = spl[1]
		}
		if len(spl) > 2 {
			return 0, 0, false
		}
	}
	var err error
	if hour, err = strconv.Atoi(hourS); err != nil {
		return 0, 0, false
	}
	if min, err = strconv.Atoi(minS); err != nil {
		return 0, 0, false
	}
	return hour, min, true
}

func (c *ctxTime) IsAfter(time string) bool {
	hr, min, ok := parseTime(time)
	if !ok {
		return false
	}
	return c.Hour() >= hr && c.Minute() >= min
}

///

func (e *ctxEvent) getProp(prop ics.ComponentProperty) string {
	if v := e.GetProperty(prop); v != nil {
		return v.Value
	}
	return ""
}

func (e *ctxEvent) Description() string {
	return e.getProp(ics.ComponentPropertyDescription)
}

func (e *ctxEvent) Summary() string {
	return e.getProp(ics.ComponentPropertySummary)
}

func (e *ctxEvent) URL() string {
	return e.getProp(ics.ComponentPropertyUrl)
}

func (e *ctxEvent) Categories() string {
	return e.getProp(ics.ComponentPropertyCategories)
}

func (e *ctxEvent) Location() string {
	return e.getProp(ics.ComponentPropertyLocation)
}

func (c *ContextFlow) CreateEnv(event *ics.VEvent) (*ContextEnv, error) {
	start, err := event.GetStartAt()
	if err != nil {
		return nil, err
	}
	ctxStart := ctxTime{&start}

	end, err := event.GetEndAt()
	if err != nil {
		return nil, err
	}
	ctxEnd := ctxTime{&end}

	ctxEv := ctxEvent{event}

	return &ContextEnv{
		Event: &ctxEv,
		Date:  &ctxStart,
		Start: &ctxStart,
		End:   &ctxEnd,
	}, nil
}
