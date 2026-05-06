package tokenizer

type Tokenizer struct {
	groups map[string]group
}

type router struct {
	varGroupName     string
	path             string
	method           string
	handlerFunc      string
	requestBodyDir   string
	requestBodyValue string
	requestBodyType  string
}

type group struct {
	VarName string
	path    string
	routers map[string]router
}

func (t *Tokenizer) editRouter(rt router) {
	rg, ok := t.groups[rt.varGroupName]
	if !ok {
		return
	}

	_, ok = rg.routers[rt.handlerFunc]
	if !ok {
		return
	}

	t.groups[rt.varGroupName].routers[rt.handlerFunc] = rt
}

func (t *Tokenizer) getRouter(groupName string, routerName string) (router, bool) {
	rg, ok := t.groups[groupName]
	if !ok {
		return router{}, false
	}

	rt, ok := rg.routers[routerName]
	if !ok {
		return router{}, false
	}

	return rt, true
}

func (t *Tokenizer) addRoute(groupName string, route router) {
	_, ok := t.groups[groupName]
	if !ok {
		return
	}
	t.groups[groupName].routers[route.handlerFunc] = route
}

func (t *Tokenizer) addGroup(group group) {
	t.groups[group.VarName] = group
}

func (t *Tokenizer) getGroup(varName string) (group, bool) {
	if g, ok := t.groups[varName]; ok {
		return g, ok
	}
	return group{}, false
}
