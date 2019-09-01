package dockercompose

import "fmt"

type Network struct {
	Name string
}

func newNetwork(name string) *Network {
	return &Network{Name: name}
}

func (n *Network) ToYML() string {
	return fmt.Sprintf("  %s:\n", n.Name)
}
