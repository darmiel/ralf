package main

import (
	"fmt"
	"github.com/ralf-life/engine/model"
	"io"
	"os"
)

type ContextFlow struct {
	*model.Profile
	Context map[string]interface{}
}

// ret -> return
// show -> show event in calendar
func (c *ContextFlow) run(flow model.Flow) (ret bool, show bool) {
	switch f := flow.(type) {
	case *model.ReturnFlow:
		ret = true
		return
	case *model.DebugFlow:
		fmt.Println("[DEBUG]", f.Debug)
	case *model.ConditionFlow:
		// check condition
		if f.Condition == "true" {
			for _, child := range f.Then {
				if r, _ := c.run(child); r {
					ret = r
					return
				}
			}
		} else {
			for _, child := range f.Else {
				if r, _ := c.run(child); r {
					ret = r
					return
				}
			}
		}
	}
	return
}

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
	f, err := os.Open("example-profile.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var profile *model.Profile
	// profile = testYaml(f)
	profile = testJson(f)

	if profile == nil {
		fmt.Println("profile was nil.")
		return
	}

	fmt.Printf("%+v\n", profile)
	cp := ContextFlow{profile, make(map[string]interface{})}

	fmt.Println()
	fmt.Println("running")
	for _, flow := range profile.Flows {
		ret, show := cp.run(flow)
		if ret {
			break
		}
		_ = show
	}
}
