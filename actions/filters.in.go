package actions

import ics "github.com/darmiel/golang-ical"

type FilterInAction struct{}

func (fia *FilterInAction) Identifier() string {
	return "filters/filter-in"
}

///

var DummyFilterInMessage = FilterInMessage(1)

func (fia *FilterInAction) Execute(_ *ics.VEvent, _ map[string]interface{}, _ bool) (ActionMessage, error) {
	return DummyFilterInMessage, nil
}
