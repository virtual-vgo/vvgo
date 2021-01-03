package api

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_rankResults(t *testing.T) {
	results := [][]string{{"Choice A", "Choice B"}, {"Choice C"}, {"Choice D"}, {"Choice E"}, {"Choice F"}}
	assert.Equal(t, map[int][]string{
		1: {"Choice A", "Choice B"},
		3: {"Choice C"},
		4: {"Choice D"},
		5: {"Choice E"},
		6: {"Choice F"},
	}, rankResults(results))
}

func Test_determineWinners(t *testing.T) {
	t.Run("no ties", func(t *testing.T) {
		ballots := [][]string{
			{"Choice A", "Choice B", "Choice C", "Choice D", "Choice E", "Choice F"},
			{"Choice D", "Choice B", "Choice A", "Choice C", "Choice E", "Choice F"},
			{"Choice C", "Choice B", "Choice E", "Choice A", "Choice D", "Choice F"},
			{"Choice C", "Choice B", "Choice D", "Choice E", "Choice A", "Choice F"},
			{"Choice A", "Choice B", "Choice C", "Choice D", "Choice E", "Choice F"},
		}
		assert.Equal(t, [][]string{
			{"Choice A"}, {"Choice B"}, {"Choice C"}, {"Choice D"}, {"Choice E"}, {"Choice F"},
		}, determineWinners(ballots))
	})

	t.Run("tie", func(t *testing.T) {
		ballots := [][]string{
			{"Choice A", "Choice B", "Choice C", "Choice D", "Choice E", "Choice F"},
			{"Choice C", "Choice B", "Choice E", "Choice A", "Choice D", "Choice F"},
			{"Choice C", "Choice B", "Choice D", "Choice E", "Choice A", "Choice F"},
			{"Choice A", "Choice B", "Choice C", "Choice D", "Choice E", "Choice F"},
		}
		assert.Equal(t, [][]string{
			{"Choice A", "Choice C"}, {"Choice B"}, {"Choice D"}, {"Choice E"}, {"Choice F"},
		}, determineWinners(ballots))
	})

	t.Run("tie with lifting", func(t *testing.T) {
		ballots := [][]string{
			{"Choice A", "Choice B", "Choice C", "Choice D", "Choice E", "Choice F"},
			{"Choice D", "Choice B", "Choice A", "Choice C", "Choice E", "Choice F"},
			{"Choice C", "Choice B", "Choice D", "Choice E", "Choice A", "Choice F"},
			{"Choice A", "Choice B", "Choice C", "Choice D", "Choice E", "Choice F"},
		}
		assert.Equal(t, [][]string{
			{"Choice A", "Choice B"}, {"Choice C"}, {"Choice D"}, {"Choice E"}, {"Choice F"},
		}, determineWinners(ballots))
	})
}
