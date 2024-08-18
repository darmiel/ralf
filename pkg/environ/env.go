package environ

import (
	"fmt"
	ics "github.com/darmiel/golang-ical"
	"github.com/darmiel/ralf/internal/util"
	"strconv"
	"strings"
	"time"
)

type ExprEnvironment struct {
	Event   CtxEvent
	Date    CtxTime
	Start   CtxTime
	End     CtxTime
	Context util.NamedValues
}

func (e *ExprEnvironment) AORB(val bool, a, b string) string {
	if val {
		return a
	}
	return b
}

func (e *ExprEnvironment) String(obj any) string {
	return fmt.Sprintf("%+v", obj)
}

func (e *ExprEnvironment) Lower(inp string) string {
	return strings.ToLower(inp)
}

func (e *ExprEnvironment) Upper(inp string) string {
	return strings.ToUpper(inp)
}

func (e *ExprEnvironment) Trim(inp string) string {
	return strings.TrimSpace(inp)
}

func (e *ExprEnvironment) Split(inp, sep string) []string {
	return strings.Split(inp, sep)
}

func (e *ExprEnvironment) Join(elems []string, sep string) string {
	return strings.Join(elems, sep)
}

func (e *ExprEnvironment) Repeat(what string, times int) string {
	return strings.Repeat(what, times)
}

func (e *ExprEnvironment) Count(str, substr string) int {
	return strings.Count(str, substr)
}

func (e *ExprEnvironment) Replace(str, old, new string) string {
	return strings.ReplaceAll(str, old, new)
}

type (
	CtxTime struct {
		time *time.Time
	}
	CtxEvent struct {
		event *ics.VEvent
	}
)

func NewTime(time time.Time) CtxTime {
	return CtxTime{time: &time}
}

func NewEvent(event ics.VEvent) CtxEvent {
	return CtxEvent{event: &event}
}

func (c CtxTime) IsMonday() bool {
	return c.time.Weekday() == time.Monday
}

func (c CtxTime) IsTuesday() bool {
	return c.time.Weekday() == time.Tuesday
}

func (c CtxTime) IsWednesday() bool {
	return c.time.Weekday() == time.Wednesday
}

func (c CtxTime) IsThursday() bool {
	return c.time.Weekday() == time.Thursday
}

func (c CtxTime) IsFriday() bool {
	return c.time.Weekday() == time.Friday
}

func (c CtxTime) IsSaturday() bool {
	return c.time.Weekday() == time.Saturday
}

func (c CtxTime) IsSunday() bool {
	return c.time.Weekday() == time.Sunday
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

func (c CtxTime) IsAfter(time string) bool {
	hr, min, ok := parseTime(time)
	if !ok {
		return false
	}
	return c.time.Hour() >= hr && c.time.Minute() >= min
}

///

func (e CtxEvent) getProp(prop ics.ComponentProperty) string {
	if v := e.event.GetProperty(prop); v != nil {
		return v.Value
	}
	return ""
}

func (e CtxEvent) Description() string {
	return e.getProp(ics.ComponentPropertyDescription)
}

func (e CtxEvent) Summary() string {
	return e.getProp(ics.ComponentPropertySummary)
}

func (e CtxEvent) URL() string {
	return e.getProp(ics.ComponentPropertyUrl)
}

func (e CtxEvent) Categories() string {
	return e.getProp(ics.ComponentPropertyCategories)
}

func (e CtxEvent) Location() string {
	return e.getProp(ics.ComponentPropertyLocation)
}

func (e CtxEvent) HasAttendee(mail string) bool {
	return util.HasAttendee(e.event, mail)
}

func CreateExprEnvironmentFromEvent(event *ics.VEvent, sharedContext util.NamedValues) (*ExprEnvironment, error) {
	start, err := event.GetStartAt()
	if err != nil {
		return nil, fmt.Errorf("get start at err: %v", err)
	}
	ctxStart := CtxTime{&start}

	end, err := event.GetEndAt()
	if err != nil {
		return nil, fmt.Errorf("get end at err: %v", err)
	}
	ctxEnd := CtxTime{&end}

	return &ExprEnvironment{
		Event: CtxEvent{
			event: event,
		},
		Date:    ctxStart,
		Start:   ctxStart,
		End:     ctxEnd,
		Context: sharedContext,
	}, nil
}
