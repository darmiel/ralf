package actions

import ics "github.com/darmiel/golang-ical"

type FilterOutAction struct{}

func (foa *FilterOutAction) Identifier() string {
	return "filters/filter-out"
}

///

func (foa *FilterOutAction) Execute(_ *ics.VEvent, _ map[string]interface{}, _ bool) (ActionMessage, error) {
	return new(FilterOutActionMessage), nil
}
