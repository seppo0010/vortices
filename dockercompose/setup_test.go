package dockercompose

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupYMLTwoNetworks(t *testing.T) {
	setup := NewSetup()
	network1 := setup.NewNetwork("network1")
	network2 := setup.NewNetwork("network2")
	setup.NewComputer("computer", "ubuntu", nil, []*Network{
		network1,
		network2,
	}).ToYML()
	assert.Equal(t, setup.ToYML(), fmt.Sprintf(`
version: "2.1"
services:
  %s_computer:
    container_name: %s_computer
    image: ubuntu
    networks:
      %s_network1:
      %s_network2:


networks:
  %s_network1:
  %s_network2:
`, setup.ID, setup.ID, setup.ID, setup.ID, setup.ID, setup.ID))
}
