package dockercompose

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type BaseComputer struct {
	setup    *Setup
	Name     string
	Image    string
	Networks []*Network
}

func (comp *BaseComputer) ToYML() string {
	networks := ""
	ports := ""
	if len(comp.Networks) > 0 {
		networks = "    networks:\n"
		for _, network := range comp.Networks {
			networks += fmt.Sprintf("      %s:\n", network.Name)
		}
	}
	return fmt.Sprintf(`  %s:
    container_name: %s
    image: %s
    privileged: true
%s
%s
`, comp.Name, comp.Name, comp.Image, networks, ports)
}

func newBaseComputer(setup *Setup, name, image string, networks []*Network) *BaseComputer {
	return &BaseComputer{
		setup:    setup,
		Name:     setup.makeName(name),
		Image:    image,
		Networks: networks,
	}
}

func newComputer(setup *Setup, name, image string, gateway *Router, networks []*Network) *Computer {
	return &Computer{
		BaseComputer: newBaseComputer(setup, name, image, networks),
		Gateway:      gateway,
	}
}

type Computer struct {
	*BaseComputer
	Gateway *Router
}

func (comp *BaseComputer) GetIPAddress() string {
	ips, err := comp.GetAllIPAddresses()
	if err != nil {
		log.Fatalf("failed to get ip addresses: %s", err.Error())
	}
	return ips[0]
}

func (comp *BaseComputer) GetIPAddressForNetwork(network *Network) (string, error) {
	networksExec := comp.setup.exec(runRequest{
		args: []string{"docker", "inspect", "-f", "{{json .NetworkSettings.Networks}}", comp.Name},
	})
	if networksExec.err != nil {
		return "", networksExec.err
	}
	var networks map[string]map[string]interface{}
	err := json.Unmarshal(networksExec.stdout, &networks)
	if err != nil {
		return "", err
	}

	for network_id, data := range networks {
		networkLabelExec := comp.setup.exec(runRequest{
			args: []string{"docker", "inspect", "-f", "{{range $key, $value := .Labels}}{{if eq $key \"com.docker.compose.network\"}}{{$value}}{{end}}{{end}}", network_id},
		})
		if networkLabelExec.err != nil {
			return "", err
		}

		if strings.Trim(string(networkLabelExec.stdout), " \n") == network.Name {
			ip, found := data["IPAddress"]
			if !found {
				return "", fmt.Errorf("ip address not found")
			}
			ipStr, ok := ip.(string)
			if !ok {
				return "", fmt.Errorf("invalid ip address, got %T", ip)
			}
			return ipStr, nil
		}
	}
	return "", fmt.Errorf("could not find ip address for %s in network %s", comp.Name, network.Name)
}

func (comp *BaseComputer) GetAllIPAddresses() ([]string, error) {
	networksExec := comp.setup.exec(runRequest{
		args: []string{"docker", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}} {{end}}", comp.Name},
	})
	if networksExec.err != nil {
		return nil, networksExec.err
	}
	return strings.Split(strings.Trim(string(networksExec.stdout), " \n"), " "), nil
}

func findSharedNetwork(networks1, networks2 []*Network) *Network {
	for _, n1 := range networks1 {
		for _, n2 := range networks2 {
			if n1 == n2 {
				return n1
			}
		}
	}
	return nil
}

func (comp *BaseComputer) GetIPAddressFor(comp2 *BaseComputer) (string, error) {
	network := findSharedNetwork(comp.Networks, comp2.Networks)
	if network == nil {
		return "", fmt.Errorf("no shared network found between %s and %s", comp.Name, comp2.Name)
	}
	return comp2.GetIPAddressForNetwork(network)
}

func (comp *BaseComputer) Start() error {
	return nil
}

func (comp *Computer) Start() error {
	if err := comp.BaseComputer.Start(); err != nil {
		return err
	}
	if comp.Gateway != nil {
		ipAddress, err := comp.GetIPAddressFor(comp.Gateway.BaseComputer)
		if err != nil {
			return err
		}
		ipRouteDelDefault := comp.setup.exec(runRequest{
			args: []string{"docker", "exec", "--privileged", comp.Name, "ip", "route", "del", "default"},
		})
		if ipRouteDelDefault.err != nil {
			return err
		}

		ipRouteAddDefault := comp.setup.exec(runRequest{
			args: []string{"docker", "exec", "--privileged", comp.Name, "ip", "route", "add", "default", "via", ipAddress},
		})
		if ipRouteAddDefault.err != nil {
			return ipRouteAddDefault.err
		}
	}
	return nil
}
