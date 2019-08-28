package dockercompose

import (
	"os/exec"
)

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

func (router *Router) Start() error {
	cmd := exec.Command("./router/start-router", router.Name)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
