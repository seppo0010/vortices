package dockercompose

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/google/uuid"
)

type Setup struct {
	ID          string
	tmpDir      string
	Computers   []*Computer
	STUNServers []*STUNServer
	Routers     []*Router
	Networks    []*Network
}

func NewSetup() *Setup {
	return &Setup{ID: uuid.New().String(), Computers: []*Computer{}, Networks: []*Network{}, Routers: []*Router{}}
}

func (s *Setup) makeName(name string) string {
	return fmt.Sprintf("%s_%s", s.ID, name)
}
func (s *Setup) NewNetwork(name string) *Network {
	network := newNetwork(s.makeName(name))
	s.Networks = append(s.Networks, network)
	return network
}

func (s *Setup) NewComputer(name, image string, gateway *Router, networks []*Network) *Computer {
	computer := newComputer(s, name, image, gateway, networks)
	s.Computers = append(s.Computers, computer)
	return computer
}

func (s *Setup) NewRouter(name, image string, networks []*Network) *Router {
	router := newRouter(s, name, image, networks)
	s.Routers = append(s.Routers, router)
	return router
}

func (s *Setup) NewSTUNServer(name string, networks []*Network) *STUNServer {
	stunServer := newSTUNServer(s, name, networks)
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
	setup.tmpDir = path.Join(os.TempDir(), "vortices", setup.ID)
	err := os.MkdirAll(setup.tmpDir, 0744)
	if err != nil {
		return err
	}
	f, err := os.Create(path.Join(setup.tmpDir, "docker-compose.yml"))
	if err != nil {
		return err
	}
	_, err = f.WriteString(setup.ToYML())
	if err != nil {
		return err
	}
	f.Close()

	cmd := setup.exec(runRequest{args: []string{"docker-compose", "up", "-d"}})
	if cmd.err != nil {
		log.Fatalf("failed to start docker-compose")
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
	cmd := setup.exec(runRequest{args: []string{"docker-compose", "down"}})
	if cmd.err != nil {
		return cmd.err
	}
	err := os.RemoveAll(setup.tmpDir)
	if err != nil {
		return err
	}
	return nil
}
