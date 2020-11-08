package api

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func TestCreditsToWebsitePasta(t *testing.T) {
	expected, err := ioutil.ReadFile("testdata/website-credits-pasta.txt")
	require.NoError(t, err, "ioutil.ReadFile() failed")
	submissionRecords := ValuesToSubmissionRecords(readValuesFromFile(t, "testdata/submission-records.csv", ','))
	credits := SubmissionRecordsToCredits("04-between-heaven-and-earth", submissionRecords)
	got := CreditsToWebsitePasta(credits)
	assert.Equal(t, string(expected), got)
}

func TestCreditsToVideoPasta(t *testing.T) {
	expected, err := ioutil.ReadFile("testdata/video-credits-pasta.txt")
	require.NoError(t, err, "ioutil.ReadFile() failed")
	submissionRecords := ValuesToSubmissionRecords(readValuesFromFile(t, "testdata/submission-records.csv", ','))
	credits := SubmissionRecordsToCredits("04-between-heaven-and-earth", submissionRecords)
	got := CreditsToVideoPasta(credits)
	assert.Equal(t, string(expected), got)
	fmt.Println(got)
}

func TestCreditsToYoutubePasta(t *testing.T) {
	expected, err := ioutil.ReadFile("testdata/youtube-credits-pasta.txt")
	require.NoError(t, err, "ioutil.ReadFile() failed")
	submissionRecords := ValuesToSubmissionRecords(readValuesFromFile(t, "testdata/submission-records.csv", ','))
	credits := SubmissionRecordsToCredits("04-between-heaven-and-earth", submissionRecords)
	got := CreditsToYoutubePasta(credits)
	assert.Equal(t, string(expected), got)
}
