package actions

type FilterOutAction struct{}

func (foa *FilterOutAction) Identifier() string {
	return "filters/filter-out"
}

///

func (foa *FilterOutAction) Execute(_ *Context) (ActionMessage, error) {
	return new(FilterOutActionMessage), nil
}
