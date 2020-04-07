package main

import "net/http"

func badRequest(w http.ResponseWriter) {
	http.Error(w, "", http.StatusBadRequest)
}

func internalServerError(w http.ResponseWriter) {
	http.Error(w, "", http.StatusInternalServerError)
}

func methodNotAllowed(w http.ResponseWriter) {
	http.Error(w, "", http.StatusMethodNotAllowed)
}

func tooManyBytes(w http.ResponseWriter) {
	http.Error(w, "", http.StatusRequestEntityTooLarge)
}

func invalidContent(w http.ResponseWriter) {
	http.Error(w, "", http.StatusUnsupportedMediaType)
}
