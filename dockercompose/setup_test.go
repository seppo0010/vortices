package dockercompose

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetupYMLTwoNetworks(t *testing.T) {
	setup := NewSetup()
	network1 := setup.NewNetwork("network1")
	network2 := setup.NewNetwork("network2")
	setup.NewComputer("computer", "ubuntu", []*Network{
		network1,
		network2,
	}).ToYML()
	assert.Equal(t, setup.ToYML(), `
version: "3"
services:
  computer:
    image: ubuntu
    networks:
      - network1
      - network2

`)
}
