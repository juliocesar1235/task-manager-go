package middleware

import "net/http"

type contextKey string

const RequestIDKey contextKey = "request_id"

func GetRequestID(r *http.Request) string {
	id, _ := r.Context().Value(RequestIDKey).(string)
	return id
}
