package dockercompose

type Router struct {
	*BaseComputer
}

func newRouter(name, image string, networkIPv4 map[string]string, networks []*Network) *Router {
	router := &Router{
		BaseComputer: newBaseComputer(name, image, networks),
	}
	router.NetworkIPv4 = networkIPv4
	return router
}
