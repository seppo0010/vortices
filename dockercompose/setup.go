package dockercompose

func NewSetup() *Setup {
	return &Setup{Computers: []*Computer{}, Networks: []*Network{}}
}

func (s *Setup) NewNetwork(name string) *Network {
	network := newNetwork(name)
	s.Networks = append(s.Networks, network)
	return network
}

func (s *Setup) NewComputer(name string, networks []*Network) *Computer {
	computer := newComputer(name, networks)
	s.Computers = append(s.Computers, computer)
	return computer
}

func (s *Setup) ToYML() string {
	yml := `
version: "3"
services:
`
	for _, comp := range s.Computers {
		yml += comp.ToYML()
	}
	return yml
}

