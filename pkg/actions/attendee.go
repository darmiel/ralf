package actions

import (
	"fmt"
	ics "github.com/darmiel/golang-ical"
	"github.com/darmiel/ralf/internal/util"
	"strings"
)

type ClearAttendeesAction struct{}

func (*ClearAttendeesAction) Identifier() string {
	return "actions/clear-attendees"
}

func (*ClearAttendeesAction) Execute(ctx *Context) (ActionMessage, error) {
	clearAttendees, err := optional[bool](ctx.With, "attendees", true)
	if err != nil {
		return nil, err
	}
	clearOrganizer, err := optional[bool](ctx.With, "organizer", false)
	if err != nil {
		return nil, err
	}
	for i := len(ctx.Event.Properties) - 1; i >= 0; i-- {
		doAttendees := clearAttendees && ctx.Event.Properties[i].IANAToken == string(ics.PropertyAttendee)
		doOrganizer := clearOrganizer && ctx.Event.Properties[i].IANAToken == string(ics.PropertyOrganizer)
		if doAttendees || doOrganizer {
			ctx.Event.Properties = append(ctx.Event.Properties[:i], ctx.Event.Properties[i+1:]...)
		}
	}
	return nil, nil
}

// ---

type AddAttendeeAction struct{}

func (*AddAttendeeAction) Identifier() string {
	return "actions/add-attendee"
}

func (*AddAttendeeAction) Execute(ctx *Context) (ActionMessage, error) {
	mail, err := required[string](ctx.With, "mail")
	if err != nil {
		return nil, err
	}

	var props []ics.PropertyParameter
	if status, _ := optional[string](ctx.With, "status", ""); status != "" {
		switch strings.ToLower(status) {
		case "needs-action":
			props = append(props, ics.ParticipationStatusNeedsAction)
		case "accepted":
			props = append(props, ics.ParticipationStatusAccepted)
		case "declined":
			props = append(props, ics.ParticipationStatusTentative)
		case "delegated":
			props = append(props, ics.ParticipationStatusDelegated)
		case "completed":
			props = append(props, ics.ParticipationStatusCompleted)
		case "in-process":
			props = append(props, ics.ParticipationStatusInProcess)
		default:
			return nil, fmt.Errorf("unknown status: %s", status)
		}
	}

	if role, _ := optional[string](ctx.With, "role", ""); role != "" {
		switch strings.ToLower(role) {
		case "chair":
			props = append(props, ics.ParticipationRoleChair)
		case "required":
			props = append(props, ics.ParticipationRoleReqParticipant)
		case "optional":
			props = append(props, ics.ParticipationRoleOptParticipant)
		case "non-participant":
			props = append(props, ics.ParticipationRoleNonParticipant)
		default:
			return nil, fmt.Errorf("unknown role: %s", role)
		}
	}

	if ctx.Verbose {
		fmt.Println("[actions/add-attendee] props:", props)
	}

	// check if event already has attendee
	if !util.HasAttendee(ctx.Event, mail) {
		ctx.Event.AddAttendee(mail, props...)
	}
	return nil, nil
}
