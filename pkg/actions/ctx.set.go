package actions

import (
	"errors"
	"fmt"
	"github.com/antonmedv/expr"
	"github.com/darmiel/ralf/internal/util"
	"github.com/darmiel/ralf/pkg/environ"
	"strings"
)

type CtxSetAction struct{}

func (c *CtxSetAction) Identifier() string {
	return "ctx/set"
}

var (
	ErrKeyInSharedContext = errors.New("key already in shared context. set #overwrite to true to overwrite")
	ErrNotString          = errors.New("dynamic values must be of type string")
)

type ctxSetExprEnv struct {
	environ.ExprEnvironment
	With util.NamedValues
}

func (c *CtxSetAction) Execute(ctx *Context) (ActionMessage, error) {
	overwrite, err := optional(ctx.With, "#overwrite", true)
	if err != nil {
		return nil, err
	}
	for k, v := range ctx.With {
		dynamic := strings.HasPrefix(k, "$")
		if dynamic {
			k = strings.TrimLeft(k, "$")
			// for dynamic values, the value must be a string.
			if _, ok := v.(string); !ok {
				return nil, ErrNotString
			}
		}
		// if already in shared context, and we don't want to overwrite, panic
		if _, ok := ctx.SharedContext[k]; ok && !overwrite {
			return nil, ErrKeyInSharedContext
		}
		if dynamic {
			defaultEnv, err := environ.CreateExprEnvironmentFromEvent(ctx.Event, ctx.SharedContext)
			if err != nil {
				return nil, err
			}
			env := ctxSetExprEnv{
				ExprEnvironment: *defaultEnv,
				With:            ctx.With,
			}
			eval, err := expr.Eval(v.(string), &env)
			if err != nil {
				return nil, err
			}
			v = eval
		}
		ctx.SharedContext[k] = v
		if ctx.Verbose {
			fmt.Printf("[ctx/set] Set (%s) to '%+v'\n", k, v)
		}
	}
	return nil, nil
}
