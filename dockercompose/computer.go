package dockercompose

import (
	"bytes"
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

func (comp *BaseComputer) GetIPAddress() string {
	ips, err := comp.GetAllIPAddresses()
	if err != nil {
		log.Fatalf("failed to get ip addresses: %s", err.Error())
	}
	return ips[0]
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

func (comp *Computer) Start() error {
	if comp.Gateway != "" {
		cmd := exec.Command("docker", "exec", "--privileged", comp.Name, "ip", "route", "del", "default")
		err := cmd.Run()
		if err != nil {
			return err
		}

		cmd = exec.Command("docker", "exec", "--privileged", comp.Name, "ip", "route", "add", "default", "via", comp.Gateway)
		err = cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}
