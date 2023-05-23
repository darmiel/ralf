package actions

import (
	"errors"
	"fmt"
	"github.com/antonmedv/expr"
)

type CtxSetAction struct{}

func (c *CtxSetAction) Identifier() string {
	return "ctx/set"
}

var (
	ErrKeyInSharedContext = errors.New("key already in shared context. set $overwrite to true to overwrite")
)

func (c *CtxSetAction) Execute(ctx *Context) (ActionMessage, error) {
	overwrite, err := optional(ctx.With, "$overwrite", false)
	if err != nil {
		return nil, err
	}
	eval, err := optional(ctx.With, "$eval", "")
	if err != nil {
		return nil, err
	}
	for k, v := range ctx.With {
		// if already in shared context, and we don't want to overwrite, panic
		if _, ok := ctx.SharedContext[k]; ok && !overwrite {
			return nil, ErrKeyInSharedContext
		}
		if eval != "" {
			if ctx.Verbose {
				fmt.Printf("evaluating $eval expression: '%s'...\n", eval)
			}
			if eval, err := expr.Eval(eval, map[string]interface{}{
				"Value": v,
			}); err != nil {
				return nil, fmt.Errorf("cannot evaluate $eval-expresison: %v", err)
			} else {
				fmt.Printf("result: '%+v'\n", v)
				v = eval
			}
		}
		ctx.SharedContext[k] = v
	}
	return nil, nil
}
