package main

import (
	"fmt"
	"github.com/ralf-life/engine/model"
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

func main() {
	f, err := os.Open("example-profile.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// parse profile "example-profile.yaml"
	profile, err := model.ParseProfileFromYAML(f)
	if err != nil {
		panic(err)
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
