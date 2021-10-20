package models

const StatusOk = "ok"
const StatusError = "error"

type ApiResponse struct {
	Status      string
	Error       *ErrorResponse      `json:"Error,omitempty"`
	Projects    []Project           `json:"Projects,omitempty"`
	Parts       []Part              `json:"Parts,omitempty"`
	Sessions    []Identity          `json:"Sessions,omitempty"`
	Spreadsheet *Spreadsheet        `json:"Spreadsheet,omitempty"`
	Dataset     []map[string]string `json:"Dataset,omitempty"`
	Identity    *Identity           `json:"Identity,omitempty"`
}

type ErrorResponse struct {
	Code  int
	Error string
}

type CreateSessionsRequest struct {
	Sessions []struct {
		Kind    string   `json:"kind"`
		Roles   []string `json:"roles"`
		Expires int      `json:"expires"`
	} `json:"sessions"`
}

type DeleteSessionsRequest struct {
	Sessions []string `json:"sessions"`
}

type Spreadsheet struct {
	SpreadsheetName string  `json:"spreadsheet_name"`
	Sheets          []Sheet `json:"sheets"`
}

type GetSpreadsheetRequest struct {
	SpreadsheetName string  `json:"spreadsheet_name"`
	Sheets          []Sheet `json:"sheets"`
}
