package dockercompose

import (
	"bytes"
	"log"
	"os"
	"os/exec"
)

type Setup struct {
	Computers   []*Computer
	STUNServers []*STUNServer
	Routers     []*Router
	Networks    []*Network
}

func NewSetup() *Setup {
	return &Setup{Computers: []*Computer{}, Networks: []*Network{}, Routers: []*Router{}}
}

func (s *Setup) NewNetwork(name, subnet string) *Network {
	network := newNetwork(name, subnet)
	s.Networks = append(s.Networks, network)
	return network
}

func (s *Setup) NewComputer(name, image string, gateway *Router, networks []*Network) *Computer {
	computer := newComputer(name, image, gateway, networks)
	s.Computers = append(s.Computers, computer)
	return computer
}

func (s *Setup) NewRouter(name, image string, networkIPv4 map[string]string, networks []*Network) *Router {
	router := newRouter(name, image, networkIPv4, networks)
	s.Routers = append(s.Routers, router)
	return router
}

func (s *Setup) NewSTUNServer(name string, networks []*Network) *STUNServer {
	stunServer := newSTUNServer(name, networks)
	s.STUNServers = append(s.STUNServers, stunServer)
	return stunServer
}

func (s *Setup) ToYML() string {
	yml := `
version: "2.1"
services:
`
	for _, comp := range s.Computers {
		yml += comp.ToYML()
	}
	for _, comp := range s.STUNServers {
		yml += comp.ToYML()
	}
	for _, comp := range s.Routers {
		yml += comp.ToYML()
	}
	yml += "networks:\n"
	for _, network := range s.Networks {
		yml += network.ToYML()
	}
	return yml
}

func (setup *Setup) Start() error {
	f, err := os.Create("docker-compose.yml")
	if err != nil {
		return err
	}
	_, err = f.WriteString(setup.ToYML())
	if err != nil {
		return err
	}
	f.Close()

	var stderr bytes.Buffer
	cmd := exec.Command("docker-compose", "up", "-d")
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		log.Fatalf("failed to start docker-compose: %s", err.Error())
	}

	for _, computer := range setup.Computers {
		err = computer.Start()
		if err != nil {
			setup.Stop()
			return err
		}
	}

	for _, router := range setup.Routers {
		err = router.Start()
		if err != nil {
			setup.Stop()
			return err
		}
	}

	return nil
}

func (setup *Setup) Stop() error {
	cmd := exec.Command("docker-compose", "down")
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
