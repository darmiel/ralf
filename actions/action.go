package actions

import ics "github.com/arran4/golang-ical"

var Actions = []Action{
	&FilterInAction{},
	&FilterOutAction{},
	&RegexReplaceAction{},
}

func Find(identifier string) Action {
	for _, act := range Actions {
		if act.Identifier() == identifier {
			return act
		}
	}
	return nil
}

type Action interface {
	Identifier() string
	Execute(event *ics.VEvent, with map[string]interface{}) (ActionMessage, error)
}

type ActionMessage interface {
}

type (
	FilterOutMessage byte
	FilterInMessage  byte
)

var ()
