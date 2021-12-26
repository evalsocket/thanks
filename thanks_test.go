package thanks

import (
	"github.com/stretchr/testify/assert"
	"testing"
)


func TestThanks(t *testing.T)  {
	c := NewReleaseClient("flyteorg")

	repo ,err := c.ListRepository("flyteorg")
	assert.Nil(t,err)
	assert.Greater(t,len(repo),1)

	release ,err := c.ListRelease("flyteorg")
	assert.Nil(t,err)
	assert.Greater(t,len(release),1)

	assert.Nil(t,c.ListContributorsStats("flyteorg","flyte"))
	assert.Nil(t,c.FilterContributors(*release[0], *release[1],"flyte"))

}
