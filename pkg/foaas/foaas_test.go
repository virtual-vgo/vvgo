package foaas

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFuckOff(t *testing.T) {
	got, err := FuckOff("jackson")
	assert.NoError(t, err)
	assert.Equal(t, "Eat a bag of fucking dicks. - jackson", got)
}
