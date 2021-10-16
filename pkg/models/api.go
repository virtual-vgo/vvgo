package models

const StatusOk = "ok"
const StatusError = "error"
const ResponseTypeError = "error"
const ResponseTypeSessions = "sessions"
const SessionResponseTypeDeleted = "deleted"

type Response struct {
	Status   string            `json:"status"`
	Type     string            `json:"type"`
	Error    *ErrorResponse    `json:"error,omitempty"`
	Sessions *SessionsResponse `json:"sessions,omitempty"`
}

type ErrorResponse struct {
	Code  int
	Error string
}

func NewErrorResponse(code int, err error) Response {
	return Response{
		Status: StatusError,
		Type:   ResponseTypeError,
		Error: &ErrorResponse{
			Code:  code,
			Error: err.Error(),
		},
	}
}

type SessionsResponse struct {
	Type string
}

type DeleteSessionRequest struct {
	Sessions []string `json:"sessions"`
}
