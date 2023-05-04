package actions

import (
	ics "github.com/darmiel/golang-ical"
	"testing"
)

type test struct {
	action  string
	event   func() *ics.VEvent
	error   bool
	message ActionMessage
	with    map[string]interface{}
	check   func(event *ics.VEvent) bool
}

func getAction(name string) (Action, bool) {
	for _, a := range Actions {
		if a.Identifier() == name {
			return a, true
		}
	}
	return nil, false
}

func TestMixed(t *testing.T) {
	cases := []*test{
		// filters/filter-in doesn't need an Event to work
		{
			action:  "filters/filter-in",
			message: DummyFilterInMessage,
		},
		// filters/filter-out doesn't need an Event to work
		{
			action:  "filters/filter-out",
			message: DummyFilterOutMessage,
		},
		{
			action:  "actions/regex-replace",
			message: nil,
			event: func() *ics.VEvent {
				event := ics.NewEvent("a")
				event.SetSummary("Hello World!")
				return event
			},
			with: map[string]interface{}{
				"match":   "Hello ",
				"replace": "",
				"in":      []interface{}{"summary"},
			},
			check: func(event *ics.VEvent) bool {
				prop := event.GetProperty(ics.ComponentPropertySummary)
				if prop == nil {
					return false
				}
				return prop.Value == "World!"
			},
		},
	}
	for i, c := range cases {
		action, exists := getAction(c.action)
		if !exists {
			t.Fatalf("cannot find action %s", c.action)
		}
		var event *ics.VEvent
		if c.event != nil {
			event = c.event()
		}
		resp, err := action.Execute(event, c.with, false)
		if err == nil && c.error {
			t.Fatalf("expected error for test %d but no returned", i+1)
		} else if err != nil && !c.error {
			t.Fatalf("got error %v for test %d but no error expected", err, i+1)
		}
		if resp != c.message {
			t.Fatalf("expected return %v but got %v", c.message, resp)
		}
		if c.check != nil && !c.check(event) {
			t.Fatalf("check failed for test %d", i+1)
		}
	}
}
