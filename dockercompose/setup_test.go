package dockercompose

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupYMLTwoNetworks(t *testing.T) {
	setup := NewSetup()
	network1 := setup.NewNetwork("network1", "1.2.3.4/5")
	network2 := setup.NewNetwork("network2", "2.3.4.5/6")
	setup.NewComputer("computer", "ubuntu", nil, []*Network{
		network1,
		network2,
	}).ToYML()
	assert.Equal(t, setup.ToYML(), `
version: "2"
services:
  computer:
    container_name: computer
    image: ubuntu
    networks:
      network1:
      network2:

networks:
  network1:
    ipam:
      config:
      - subnet: 1.2.3.4/5
  network2:
    ipam:
      config:
      - subnet: 2.3.4.5/6
`)
}
