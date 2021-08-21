package when2meet

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRangeDates(t *testing.T) {
	dates, err := rangeDates("2021-02-09", "2021-02-11")
	assert.NoError(t, err)
	assert.Equal(t, []string{"2021-02-09", "2021-02-10", "2021-02-11"}, dates)
}

func TestGetWhen2MeetURL(t *testing.T) {
	url, err := CreateEvent("cheesus", "2030-02-09", "2030-02-11")
	assert.NoError(t, err)
	assert.NotEmpty(t, url)
}
