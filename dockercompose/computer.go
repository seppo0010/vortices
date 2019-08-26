package dockercompose

import "fmt"

type Computer struct {
	Name     string
	Image    string
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
    container_name: %s
    image: %s
%s
`, comp.Name, comp.Name, comp.Image, networks)
}

func newComputer(name, image string, networks []*Network) *Computer {
	return &Computer{
		Name:     name,
		Image:    image,
		Networks: networks,
	}
}
