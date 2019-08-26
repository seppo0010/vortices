package dockercompose

import "fmt"

type BaseComputer struct {
	Name        string
	Image       string
	Networks    []*Network
	NetworkIPv4 map[string]string
}

func (comp *BaseComputer) ToYML() string {
	networks := ""
	if len(comp.Networks) > 0 {
		networks = "    networks:\n"
		for _, network := range comp.Networks {
			networks += fmt.Sprintf("      %s:\n", network.Name)
			if ipv4, found := comp.NetworkIPv4[network.Name]; found {
				networks += fmt.Sprintf("        ipv4_address: %s\n", ipv4)
			}
		}
	}
	return fmt.Sprintf(`  %s:
    container_name: %s
    image: %s
%s
`, comp.Name, comp.Name, comp.Image, networks)
}

func newBaseComputer(name, image string, networks []*Network) *BaseComputer {
	return &BaseComputer{
		Name:     name,
		Image:    image,
		Networks: networks,
	}
}

func newComputer(name, image, gateway string, networks []*Network) *Computer {
	return &Computer{
		BaseComputer: newBaseComputer(name, image, networks),
		Gateway:      gateway,
	}
}

type Computer struct {
	*BaseComputer
	Gateway string
}
