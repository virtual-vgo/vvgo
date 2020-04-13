package data

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidPartNames(t *testing.T) {
	partNames := ValidPartNames()
	assert.NotNil(t, partNames, "partNames")
	assert.False(t, len(partNames) == 0, "len(partNames)")
}
