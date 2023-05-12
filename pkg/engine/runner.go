package engine

import (
	ics "github.com/darmiel/golang-ical"
	"github.com/ralf-life/engine/pkg/actions"
	"github.com/ralf-life/engine/pkg/model"
)

func ModifyCalendar(ctx *ContextFlow, flows model.Flows, cal *ics.Calendar) error {
	// get components from calendar (events) and copy to slice for later modifications
	cc := cal.Components[:]

	// start from behind so we can remove from slice
	for i := len(cc) - 1; i >= 0; i-- {
		event, ok := cc[i].(*ics.VEvent)
		if !ok {
			continue
		}
		var fact actions.ActionMessage
		fact, err := ctx.RunMultiFlows(event, flows)
		if err != nil && err != ErrExited {
			return err
		}
		switch fact.(type) {
		case actions.FilterOutActionMessage, *actions.FilterOutActionMessage:
			cc = append(cc[:i], cc[i+1:]...) // remove event from components
		}
	}

	cal.Components = cc
	return nil
}
