package sheets

import (
	"context"
	"fmt"
	"google.golang.org/api/sheets/v4"
)

func ReadSheet(ctx context.Context, spreadsheetName string, sheetName string) ([][]interface{}, error) {
	srv, err := sheets.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Sheets client: %w", err)
	}

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetName, sheetName).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve data from sheet: %w", err)
	}
	return resp.Values, nil
}

func WriteSheet(ctx context.Context, spreadsheetName string, name string, values [][]interface{}) error {
	srv, err := sheets.NewService(ctx)
	if err != nil {
		return fmt.Errorf("sheets.NewService(): %w", err)
	}
	_, err = srv.Spreadsheets.Values.
		Update(spreadsheetName, name, &sheets.ValueRange{Values: values, MajorDimension: "ROWS"}).
		ValueInputOption("USER_ENTERED").
		Do()
	if err != nil {
		return fmt.Errorf("sheets.Values.Update() failed: %w", err)
	}
	return nil
}
