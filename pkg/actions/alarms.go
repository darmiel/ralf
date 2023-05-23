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

func (*ClearAlarmsAction) Execute(ctx *Context) (ActionMessage, error) {
	for i := len(ctx.Event.Properties) - 1; i >= 0; i-- {
		if ctx.Event.Properties[i].IANAToken == string(ics.ComponentVAlarm) {
			ctx.Event.Properties = append(ctx.Event.Properties[:i], ctx.Event.Properties[i+1:]...)
		}
	}
	return nil, nil
}

// ---

type AddAlarmAction struct{}

func (*AddAlarmAction) Identifier() string {
	return "actions/add-alarm"
}

func (*AddAlarmAction) Execute(ctx *Context) (ActionMessage, error) {
	action, err := required[string](ctx.With, "action")
	if err != nil {
		return nil, err
	}
	trigger, err := required[string](ctx.With, "trigger")
	if err != nil {
		return nil, err
	}

	duration, _ := optional[string](ctx.With, "duration", "")
	repeat, _ := optional[string](ctx.With, "repeat", "")
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

	alarm := ctx.Event.AddAlarm()
	alarm.SetAction(icsAction)
	alarm.SetTrigger(trigger)

	if duration != "" {
		alarm.SetProperty(ics.ComponentProperty(ics.PropertyRepeat), repeat)
		alarm.SetProperty(ics.ComponentProperty(ics.PropertyDuration), duration)
	}

	return nil, nil
}
