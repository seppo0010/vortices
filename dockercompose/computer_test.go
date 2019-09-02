package dockercompose

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputerYMLTwoNetworks(t *testing.T) {
	setup := NewSetup()
	yml := newComputer(setup, "computer", "ubuntu", nil, []*Network{
		newNetwork("network1"),
		newNetwork("network2"),
	}).ToYML()
	assert.Equal(t, yml, fmt.Sprintf(`  %s_computer:
    container_name: %s_computer
    image: ubuntu
    networks:
      network1:
      network2:


`, setup.ID, setup.ID))
}

func TestGetAllIPAddresses(t *testing.T) {
	setup := NewSetup()
	image, err := BuildDocker("noop ubuntu", `
FROM ubuntu
CMD ["sleep", "infinity"]
	`)
	if !assert.Nil(t, err, "failed to create image") {
		return
	}
	computer := setup.NewComputer("computer", image, nil, []*Network{
		setup.NewNetwork("network1"),
		setup.NewNetwork("network2"),
	})
	setup.Start()
	ips, err := computer.GetAllIPAddresses()
	assert.Nil(t, err)
	assert.Equal(t, len(ips), 2, "expected 2 IP addresses")
	for _, ip := range ips {
		ping := setup.exec(runRequest{args: []string{"ping", ip, "-c", "1", "-w", "1", "-q"}})
		if !assert.Nil(t, ping.err) {
			return
		}
	}
	setup.Stop()
	for _, ip := range ips {
		ping := setup.exec(runRequest{args: []string{"ping", ip, "-c", "1", "-w", "1", "-q"}})
		if !assert.NotNil(t, ping.err) {
			return
		}
	}
}
