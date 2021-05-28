package foaas

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestFuckOff(t *testing.T) {
	got, err := FuckOff("jackson")
	assert.NoError(t, err)
	assert.NotEmpty(t, got)
	assert.True(t, strings.HasSuffix(got, "- jackson"))
}
