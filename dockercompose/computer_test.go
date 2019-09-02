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
