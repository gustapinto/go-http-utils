package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const (
	_requestIdHeader  = "X-RequestID"
	_requestIdContext = "requestID"
)

type RequestID struct{}

func (_ *RequestID) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(_requestIdHeader)
		if len(requestID) == 0 {
			newID, err := uuid.NewV7()
			if err == nil {
				requestID = newID.String()
			}
		}

		newCtx := context.WithValue(r.Context(), _requestIdContext, requestID)

		w.Header().Add(_requestIdHeader, requestID)

		next.ServeHTTP(w, r.WithContext(newCtx))
	}
}
