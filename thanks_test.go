package thanks

import (
	"github.com/stretchr/testify/assert"
	"testing"
)


func TestThanks(t *testing.T)  {
	c := NewReleaseClient("flyteorg","flyte")
	l ,err := c.Thanks(true)
	assert.Nil(t,err)
	assert.Greater(t,len(l),1)
}
