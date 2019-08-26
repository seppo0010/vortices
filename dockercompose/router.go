package dockercompose

type Router struct {
	*BaseComputer
}

func newRouter(name string, networks []*Network) *Router {
	return &Router{
		BaseComputer: newBaseComputer(name, "ubuntu", networks),
	}
}
