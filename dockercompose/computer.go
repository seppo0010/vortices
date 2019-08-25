package dockercompose

import "fmt"

type Computer struct {
	Name     string
	Networks []*Network
}

func (comp *Computer) ToYML() string {
    networks := ""
    if len(comp.Networks) > 0 {
        networks = "    networks:\n"
        for _, network := range comp.Networks {
            networks += fmt.Sprintf("      - %s\n", network.Name)
        }
    }
	return fmt.Sprintf(`  %s:
    image: ubuntu
%s
`, comp.Name, networks)
}

func newComputer(name string, networks []*Network) *Computer {
	return &Computer{
		Name:     name,
		Networks: networks,
	}
}
