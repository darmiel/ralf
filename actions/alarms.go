package actions

import (
	"fmt"
	ics "github.com/darmiel/golang-ical"
	"strings"
)

type ClearAlarmsAction struct{}

func (*ClearAlarmsAction) Identifier() string {
	return "actions/clear-alarms"
}

func (*ClearAlarmsAction) Execute(
	event *ics.VEvent,
	_ map[string]interface{},
	_ bool,
) (ActionMessage, error) {
	for i := len(event.Properties) - 1; i >= 0; i-- {
		if event.Properties[i].IANAToken == string(ics.ComponentVAlarm) {
			event.Properties = append(event.Properties[:i], event.Properties[i+1:]...)
		}
	}
	return nil, nil
}

// ---

type AddAlarmAction struct{}

func (*AddAlarmAction) Identifier() string {
	return "actions/add-alarm"
}

func (*AddAlarmAction) Execute(
	event *ics.VEvent,
	with map[string]interface{},
	verbose bool,
) (ActionMessage, error) {
	action, err := required[string](with, "action")
	if err != nil {
		return nil, err
	}
	trigger, err := required[string](with, "trigger")
	if err != nil {
		return nil, err
	}

	duration, _ := optional[string](with, "duration", "")
	repeat, _ := optional[string](with, "repeat", "")
	// https://www.rfc-editor.org/rfc/rfc5545#section-3.6.6
	// according to rfc5545, duration and repeat are optional but both should not be empty if either of those occurs
	if (duration == "") != (repeat == "") {
		return nil, fmt.Errorf("if duration is set, repeat should also be set")
	}

	var icsAction ics.Action
	switch strings.ToLower(action) {
	case "audio":
		icsAction = ics.ActionAudio
	case "display":
		icsAction = ics.ActionDisplay
	case "email":
		icsAction = ics.ActionEmail
	case "procedure":
		icsAction = ics.ActionProcedure
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}

	alarm := event.AddAlarm()
	alarm.SetAction(icsAction)
	alarm.SetTrigger(trigger)

	if duration != "" {
		alarm.SetProperty(ics.ComponentProperty(ics.PropertyRepeat), repeat)
		alarm.SetProperty(ics.ComponentProperty(ics.PropertyDuration), duration)
	}

	return nil, nil
}
