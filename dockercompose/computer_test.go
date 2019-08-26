package dockercompose

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputerYMLTwoNetworks(t *testing.T) {
	yml := newComputer("computer", "ubuntu", []*Network{
		newNetwork("network1"),
		newNetwork("network2"),
	}).ToYML()
	assert.Equal(t, yml, `  computer:
    image: ubuntu
    networks:
      - network1
      - network2

`)
}
