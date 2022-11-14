package actions

import (
	ics "github.com/arran4/golang-ical"
)

type RegexReplaceAction struct{}

func (rra *RegexReplaceAction) Identifier() string {
	return "actions/regex-replace"
}

///

func (rra *RegexReplaceAction) Execute(event *ics.VEvent, with map[string]interface{}) (ActionMessage, error) {
	return nil, nil
}
