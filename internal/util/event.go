package util

import (
	ics "github.com/darmiel/golang-ical"
	"strings"
)

func HasAttendee(event *ics.VEvent, mail string) bool {
	for _, a := range event.Attendees() {
		if strings.ToLower(a.Email()) == strings.ToLower(mail) {
			return true
		}
	}
	return false
}
