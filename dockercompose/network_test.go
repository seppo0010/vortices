package dockercompose

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNetworkYML(t *testing.T) {
	yml := newNetwork("network").ToYML()
	assert.Equal(t, yml, `  network:
`)
}
