package dockercompose

import "fmt"

type Setup struct {
	Computers []*Computer
	Routers   []*Router
	Networks  []*Network
}

func NewSetup() *Setup {
	return &Setup{Computers: []*Computer{}, Networks: []*Network{}, Routers: []*Router{}}
}

func (s *Setup) NewNetwork(name string) *Network {
	network := newNetwork(name)
	s.Networks = append(s.Networks, network)
	return network
}

func (s *Setup) NewComputer(name, image, gateway string, networks []*Network) *Computer {
	computer := newComputer(name, image, gateway, networks)
	s.Computers = append(s.Computers, computer)
	return computer
}

func (s *Setup) ToYML() string {
	yml := `
version: "2"
services:
`
	for _, comp := range s.Computers {
		yml += comp.ToYML()
	}
	yml += "networks:\n"
	for _, network := range s.Networks {
		yml += fmt.Sprintf("  %s:\n", network.Name)
	}
	return yml
}
