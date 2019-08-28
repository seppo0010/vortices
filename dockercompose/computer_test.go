package dockercompose

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputerYMLTwoNetworks(t *testing.T) {
	yml := newComputer("computer", "ubuntu", "", []*Network{
		newNetwork("network1", "1.2.3.4/5"),
		newNetwork("network2", "2.3.4.5/6"),
	}).ToYML()
	assert.Equal(t, yml, `  computer:
    container_name: computer
    image: ubuntu
    networks:
      network1:
      network2:

`)
}
