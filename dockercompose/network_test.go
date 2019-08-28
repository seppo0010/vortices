package dockercompose

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNetworkYML(t *testing.T) {
	yml := newNetwork("network", "1.2.3.4/5").ToYML()
	assert.Equal(t, yml, `  network:
    internal: true
    ipam:
      config:
      - subnet: 1.2.3.4/5
`)
}
