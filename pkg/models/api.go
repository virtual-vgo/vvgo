package models

const StatusOk = "ok"
const StatusError = "error"

type ApiResponse struct {
	Status      string         `json:"status"`
	Error       *ErrorResponse `json:"error,omitempty"`
	Projects    []Project      `json:"projects,omitempty"`
	Parts       []Part         `json:"parts,omitempty"`
	Directors   []Director     `json:"directors,omitempty"`
	Sessions    []Identity     `json:"sessions,omitempty"`
	Spreadsheet *Spreadsheet   `json:"spreadsheet,omitempty"`
	Identity    *Identity      `json:"identity,omitempty"`
}

type ErrorResponse struct {
	Code  int
	Error string
}

type ProjectsResponse struct {
	Total    int `json:"total"`
	Projects `json:"projects"`
}

type PartsResponse struct {
	Total int    `json:"total"`
	Parts []Part `json:"parts"`
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

type CreditsResponse struct {
	Table CreditsTable `json:"table"`
}
type Spreadsheet struct {
	SpreadsheetName string  `json:"spreadsheet_name"`
	Sheets          []Sheet `json:"sheets"`
}

type GetSpreadsheetRequest struct {
	SpreadsheetName string  `json:"spreadsheet_name"`
	Sheets          []Sheet `json:"sheets"`
}
