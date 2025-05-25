package handler

import (
	"context"
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, r *http.Request, status int, body any) {
	ctx := context.WithValue(r.Context(), "status", status)
	*r = *(r.WithContext(ctx))

	w.WriteHeader(status)

	if body != nil {
		if bodyBytes, err := json.Marshal(body); err == nil {
			w.Header().Add("Content-Type", "application/json")
			w.Write(bodyBytes)
		}
	}
}
