package when2meet

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/error_wrappers"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const TimeZone = "America/Los_Angeles"

var Endpoint = "https://www.when2meet.com"
var locationReg = regexp.MustCompile(`<body onload="window\.location='(\/\?.*)'">`)

func rangeDates(startDate, endDate string) ([]string, error) {
	layout := "2006-01-02"
	start, err := time.Parse(layout, startDate)
	if err != nil {
		return nil, fmt.Errorf("time.Parse() failed: %w", err)
	}
	end, err := time.Parse(layout, endDate)
	if err != nil {
		return nil, fmt.Errorf("time.Parse() failed: %w", err)
	}

	fmt.Println(start, end)
	var dates []string
	for ; start.Before(end); start = start.Add(24 * 3600 * time.Second) {
		dates = append(dates, start.Format(layout))
	}
	dates = append(dates, end.Format(layout))
	return dates, nil
}

func CreateEvent(name, startDate, endDate string) (string, error) {
	dates, err := rangeDates(startDate, endDate)
	if err != nil {
		return "", err
	}

	data := make(url.Values)
	data.Set("NewEventName", name)
	data.Set("DateTypes", "SpecificDates")
	data.Set("PossibleDates", strings.Join(dates, "|"))
	data.Set("TimeZone", TimeZone)
	data.Set("NoEarlierThan", "14")
	data.Set("NoLaterThan", "0")
	resp, err := http.PostForm(Endpoint+"/SaveNewEvent.php", data)
	if err != nil {
		return "", error_wrappers.HTTPDoFailed(err)
	} else if resp.StatusCode != http.StatusOK {
		return "", error_wrappers.Non200StatusCode()
	}

	var body bytes.Buffer
	_, _ = body.ReadFrom(resp.Body)
	matches := locationReg.FindSubmatch(body.Bytes())
	if len(matches) != 2 {
		return "", errors.New("failed to parse response body")
	}
	return "https://when2meet.com" + string(matches[1]), nil
}
