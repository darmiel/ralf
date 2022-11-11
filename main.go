package main

import (
	"fmt"
	"github.com/ralf-life/engine/model"
	"github.com/ralf-life/engine/parse"
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
		fmt.Println("[DEBUG]", f.Message)
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
	// parse profile "example-profile.yaml"
	profile, err := parse.ParseProfile("example-profile.yaml")
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
