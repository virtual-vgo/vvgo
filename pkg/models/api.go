package models

const StatusOk = "ok"
const StatusError = "error"

type ApiResponse struct {
	Status      string         `json:"status"`
	Error       *ErrorResponse `json:"error,omitempty"`
	Projects    []Project      `json:"projects,omitempty"`
	Parts       []Part         `json:"parts,omitempty"`
	Sessions    []Identity     `json:"sessions,omitempty"`
	Spreadsheet *Spreadsheet   `json:"spreadsheet,omitempty"`
	Dataset     *Dataset       `json:"dataset,omitempty"`
	Identity    *Identity      `json:"identity,omitempty"`
}

type Dataset struct {
	Name string
	Rows []map[string]string
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
