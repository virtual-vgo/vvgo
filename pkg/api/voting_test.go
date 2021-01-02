package api

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCountBallots(t *testing.T) {
	ballots := [][]string{
		{"Choice A", "Choice B", "Choice C", "Choice D", "Choice E"},
		{"Choice D", "Choice B", "Choice A", "Choice C", "Choice E"},
		{"Choice C", "Choice B", "Choice E", "Choice A", "Choice D"},
		{"Choice C", "Choice B", "Choice D", "Choice E", "Choice A"},
		{"Choice A", "Choice B", "Choice C", "Choice D", "Choice E"},
	}
	assert.Equal(t, []string{"Choice A", "Choice B", "Choice C", "Choice D", "Choice E"}, countBallots(ballots))
}
