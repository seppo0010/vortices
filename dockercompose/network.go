package dockercompose

import "fmt"

type Network struct {
	Name   string
	Subnet string
}

func newNetwork(name, subnet string) *Network {
	return &Network{Name: name, Subnet: subnet}
}

func (n *Network) ToYML() string {
	return fmt.Sprintf("  %s:\n    ipam:\n      config:\n      - subnet: %s\n", n.Name, n.Subnet)
}
