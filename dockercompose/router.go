package dockercompose

type Router struct {
	*BaseComputer
}

func newRouter(setup *Setup, name, image string, networks []*Network) *Router {
	router := &Router{
		BaseComputer: newBaseComputer(setup, name, image, networks),
	}
	return router
}

func (router *Router) Start() error {
	for _, rr := range []runRequest{
		runRequest{args: []string{"docker", "exec", "--privileged", router.Name, "iptables", "-A", "FORWARD", "-i", "eth0", "-o", "eth1", "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT"}},
		runRequest{args: []string{"docker", "exec", "--privileged", router.Name, "iptables", "-A", "FORWARD", "-i", "eth1", "-o", "eth0", "-j", "ACCEPT"}},
		runRequest{args: []string{"docker", "exec", "--privileged", router.Name, "iptables", "-t", "nat", "-A", "POSTROUTING", "-o", "eth0", "-j", "MASQUERADE"}},
	} {
		cmd := router.setup.exec(rr)
		if cmd.err != nil {
			return cmd.err
		}
	}
	return nil
}
