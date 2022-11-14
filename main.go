package main

import (
	"fmt"
	"github.com/ralf-life/engine/actions"
	"github.com/ralf-life/engine/engine"
	"github.com/ralf-life/engine/model"
	"io"
	"os"
)

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

func main() {
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

	fmt.Printf("%+v\n", profile)
	cp := engine.ContextFlow{Profile: profile, Context: make(map[string]interface{})}

	var fact actions.ActionMessage

	fmt.Println("--------------------------------------")
	fact, err = cp.RunAllFlows(nil, profile.Flows)
	fmt.Println("--------------------------------------")

	if err != nil {
		if err == engine.ErrExited {
			fmt.Println("--> flows exited because of a return statement.")
		} else {
			fmt.Println("!!> flows failed:", err)
		}
	}

	switch fact.(type) {
	case actions.FilterInMessage:
		fmt.Println("--> FILTER IN")
	case actions.FilterOutMessage:
		fmt.Println("--> FILTER OUT")
	}
}
