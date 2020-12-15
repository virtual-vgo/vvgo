package config

import (
	"context"
	"github.com/virtual-vgo/vvgo/pkg/redis"
)

func WebsiteDataSpreadsheetID(ctx context.Context) string {
	var spreadsheetID string
	redis.Do(ctx, redis.Cmd(&spreadsheetID, "GET", "config:website_data_spreadsheet_id"))
	return spreadsheetID
}

func DistroBucket(ctx context.Context) string {
	var distroBucket string
	redis.Do(ctx, redis.Cmd(&distroBucket, "GET", "config:distro_bucket"))
	return distroBucket
}
