package main

import (
	"fmt"
	ics "github.com/arran4/golang-ical"
	"github.com/ralf-life/engine/actions"
	"github.com/ralf-life/engine/engine"
	"github.com/ralf-life/engine/model"
	"io"
	"os"
)

func main() {
	localTest()
}

///

func testYaml(reader io.Reader) *model.Profile {
	// parse profile "example-profile.yaml"
	profile, err := model.ParseProfileFromYAML(reader)
	if err != nil {
		panic(err)
	}
	return profile
}

func testJson(reader io.Reader) *model.Profile {
	data, err := io.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	profile, err := model.ParseProfileFromJSON(data)
	if err != nil {
		panic(err)
	}
	return profile
}

func localTest() {
	f, err := os.Open("example-profile.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var profile *model.Profile
	profile = testYaml(f)
	// profile = testJson(f)

	if profile == nil {
		fmt.Println("profile was nil.")
		return
	}

	// read ics file
	cf, err := os.Open("TINF21B2.ics")
	if err != nil {
		panic(err)
	}
	defer cf.Close()
	cal, err := ics.ParseCalendar(cf)
	if err != nil {
		panic(err)
	}
	cal.SetMethod(ics.MethodRequest)

	cp := engine.ContextFlow{Profile: profile, Context: make(map[string]interface{})}

	// get components from calendar (events) and copy to slice for later modifications
	cc := cal.Components[:]

	// start from behind so we can remove from slice
	for i := len(cc) - 1; i >= 0; i-- {
		event, ok := cc[i].(*ics.VEvent)
		if !ok {
			continue
		}
		var fact actions.ActionMessage
		fact, err = cp.RunAllFlows(event, profile.Flows)
		if err != nil {
			if err == engine.ErrExited {
				fmt.Println("--> flows exited because of a return statement.")
			} else {
				fmt.Println("!!> flows failed:", err)
			}
		}
		switch fact.(type) {
		case actions.FilterOutMessage:
			cc = append(cc[:i], cc[i+1:]...) // remove event from components
			fmt.Println("--> FILTER OUT")
		}
	}

	cal.Components = cc
	// fmt.Println(cal.Serialize())
}
