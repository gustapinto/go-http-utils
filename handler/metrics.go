package handler

import (
	"net/http"
)

func Alive(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, http.StatusOK, map[string]any{
		"alive": true,
	})
}
