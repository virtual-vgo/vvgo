package models

import "time"

const StatusOk = "ok"
const StatusError = "error"
const ResponseTypeError = "error"
const ResponseTypeProjects = "projects"
const ResponseTypeCredits = "credits"
const ResponseTypeParts = "parts"
const ResponseTypeDirectors = "directors"
const ResponseTypeSessions = "sessions"
const SessionResponseTypeDeleted = "deleted"
const ResponseTypeSpreadsheet = "spreadsheet"

type ApiResponse struct {
	Status      string               `json:"status"`
	Type        string               `json:"type"`
	Error       *ErrorResponse       `json:"error,omitempty"`
	Projects    *ProjectsResponse    `json:"projects,omitempty"`
	Parts       *PartsResponse       `json:"parts,omitempty"`
	Directors   *DirectorsResponse   `json:"directors,omitempty"`
	Sessions    *SessionsResponse    `json:"sessions,omitempty"`
	Spreadsheet *SpreadsheetResponse `json:"spreadsheet,omitempty"`
}

type ErrorResponse struct {
	Code  int
	Error string
}

func NewErrorResponse(code int, err error) ApiResponse {
	return ApiResponse{
		Status: StatusError,
		Type:   ResponseTypeError,
		Error: &ErrorResponse{
			Code:  code,
			Error: err.Error(),
		},
	}
}

type ProjectsResponse struct {
	Total    int `json:"total"`
	Projects `json:"projects"`
}

func NewProjectsResponse(projects Projects) ApiResponse {
	return ApiResponse{
		Status: StatusOk,
		Type:   ResponseTypeProjects,
		Projects: &ProjectsResponse{
			Total:    len(projects),
			Projects: projects,
		},
	}
}

type PartsResponse struct {
	Total int    `json:"total"`
	Parts []Part `json:"parts"`
}

func NewPartsResponse(parts []Part) ApiResponse {
	return ApiResponse{
		Status: StatusOk,
		Type:   ResponseTypeSessions,
		Parts: &PartsResponse{
			Total: len(parts),
			Parts: parts,
		},
	}
}

type DirectorsResponse struct {
	Total     int        `json:"total"`
	Directors []Director `json:"directors"`
}

func NewDirectorsResponse(directors []Director) ApiResponse {
	return ApiResponse{
		Status: StatusOk,
		Type:   ResponseTypeSessions,
		Directors: &DirectorsResponse{
			Total:     len(directors),
			Directors: directors,
		},
	}
}

type CreateSessionsRequest struct {
	Sessions []struct {
		Kind    string        `json:"kind"`
		Roles   []string      `json:"roles"`
		Expires time.Duration `json:"expires"`
	} `json:"sessions"`
}

type DeleteSessionsRequest struct {
	Sessions []string `json:"sessions"`
}

type SessionsResponse struct {
	Total    int        `json:"total"`
	Sessions []Identity `json:"sessions"`
}

func NewSessionsResponse(sessions []Identity) ApiResponse {
	return ApiResponse{
		Status: StatusOk,
		Type:   ResponseTypeSessions,
		Sessions: &SessionsResponse{
			Total:    len(sessions),
			Sessions: sessions,
		},
	}
}

type CreditsResponse struct {
	Table CreditsTable `json:"table"`
}
type SpreadsheetResponse struct {
	SpreadsheetName string  `json:"spreadsheet_name"`
	Sheets          []Sheet `json:"sheets"`
}

type GetSpreadsheetRequest struct {
	SpreadsheetName string  `json:"spreadsheet_name"`
	Sheets          []Sheet `json:"sheets"`
}
