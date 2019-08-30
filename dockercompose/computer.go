package dockercompose

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type BaseComputer struct {
	Name        string
	Image       string
	Networks    []*Network
	NetworkIPv4 map[string]string
}

func (comp *BaseComputer) ToYML() string {
	networks := ""
	ports := ""
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
%s
`, comp.Name, comp.Name, comp.Image, networks, ports)
}

func newBaseComputer(name, image string, networks []*Network) *BaseComputer {
	return &BaseComputer{
		Name:        name,
		Image:       image,
		Networks:    networks,
		NetworkIPv4: map[string]string{},
	}
}

func newComputer(name, image string, gateway *Router, networks []*Network) *Computer {
	return &Computer{
		BaseComputer: newBaseComputer(name, image, networks),
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
	cmd := exec.Command("docker", "inspect", "-f", "{{json .NetworkSettings.Networks}}", comp.Name)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	var networks map[string]map[string]interface{}
	err = json.Unmarshal(stdout.Bytes(), &networks)
	if err != nil {
		return "", err
	}

	for network_id, data := range networks {
		cmd := exec.Command("docker", "inspect", "-f", "{{range $key, $value := .Labels}}{{if eq $key \"com.docker.compose.network\"}}{{$value}}{{end}}{{end}}", network_id)
		var stdout bytes.Buffer
		cmd.Stdout = &stdout
		err := cmd.Run()
		if err != nil {
			return "", err
		}

		if strings.Trim(string(stdout.Bytes()), " \n") == network.Name {
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
	cmd := exec.Command("docker", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}} {{end}}", comp.Name)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.Trim(string(stdout.Bytes()), " \n"), " "), nil
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

func (comp *Computer) Start() error {
	if comp.Gateway != nil {
		ipAddress, err := comp.GetIPAddressFor(comp.Gateway.BaseComputer)
		if err != nil {
			return err
		}
		cmd := exec.Command("docker", "exec", "--privileged", comp.Name, "ip", "route", "del", "default")
		err = cmd.Run()
		if err != nil {
			return err
		}

		cmd = exec.Command("docker", "exec", "--privileged", comp.Name, "ip", "route", "add", "default", "via", ipAddress)
		err = cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}
