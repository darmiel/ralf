package actions

import (
	ics "github.com/arran4/golang-ical"
)

type FilterInAction struct{}

func (fia *FilterInAction) Identifier() string {
	return "filters/filter-in"
}

///

var DummyFilterInMessage = FilterInMessage(1)

func (fia *FilterInAction) Execute(_ *ics.VEvent, _ map[string]interface{}) (ActionMessage, error) {
	return DummyFilterInMessage, nil
}
