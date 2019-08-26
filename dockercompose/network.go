package dockercompose

type Network struct {
	Name string
}

func newNetwork(name string) *Network {
	return &Network{Name: name}
}
