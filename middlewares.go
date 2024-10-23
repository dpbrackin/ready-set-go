package main

import (
	"log"
	"net/http"
	"time"
)

type ResponseWritter struct {
	http.ResponseWriter
	statusCode int
}

func (w *ResponseWritter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	w.statusCode = code
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := r.URL.Path

		responseWriter := &ResponseWritter{
			ResponseWriter: w,
		}

		next.ServeHTTP(responseWriter, r)

		log.Printf("[%d] %v %v", responseWriter.statusCode, path, time.Since(start))

	})
}
