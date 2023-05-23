package actions

type FilterInAction struct{}

func (fia *FilterInAction) Identifier() string {
	return "filters/filter-in"
}

///

func (fia *FilterInAction) Execute(_ *Context) (ActionMessage, error) {
	return new(FilterInActionMessage), nil
}
