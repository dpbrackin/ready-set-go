package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/dpbrackin/ready-set-go/auth"
	"github.com/dpbrackin/ready-set-go/router"
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

// AuthMidleware checks for a loged in user and passes it into the context.
// If their is no logged in user, it will reject the request with a 401.
func AuthMiddleware(srv *auth.AuthService) router.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionID, err := r.Cookie("sessionID")

			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(err.Error()))
				return
			}

			user, err := srv.AuthenticateSession(r.Context(), sessionID.Value)

			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(err.Error()))
				return
			}

			ctx := context.WithValue(r.Context(), "user", user)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
