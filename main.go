package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dpbrackin/ready-set-go/auth"
	"github.com/dpbrackin/ready-set-go/db/generated"
	"github.com/dpbrackin/ready-set-go/db/repositories"
	"github.com/dpbrackin/ready-set-go/router"
	"github.com/jackc/pgx/v5"
)

type RealClock struct{}

func (r *RealClock) Now() time.Time {
	return time.Now()
}

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DB_CONN"))

	if err != nil {
		log.Fatal(err)
		return
	}

	q := generated.New(conn)

	authService := auth.NewAuthService(auth.NewAuthServiceParams{
		Repository: repositories.NewPGAuthRepository(q),
		Clock:      &RealClock{},
	})

	authHandlers := &AuthHandlers{
		Srv: authService,
	}

	root := router.NewRootRouter()
	root.Use(LoggingMiddleware)

	unauthenticatedGroup := root.Group("")
	unauthenticatedGroup.RouteFunc("POST /login", authHandlers.Login)
	unauthenticatedGroup.RouteFunc("POST /register", authHandlers.Register)

	authenticatedGroup := root.Group("")
	authenticatedGroup.Use(AuthMiddleware(authService))
	authenticatedGroup.RouteFunc("GET /logout", authHandlers.Logout)
	authenticatedGroup.RouteFunc("GET /whoami", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(r.Context().Value("user"))
	})

	addr := ":3000"

	log.Printf("Listening on %s", addr)

	err = http.ListenAndServe(addr, root.Mux())

	if err != nil {
		log.Fatal(err)
	}

}

type AuthHandlers struct {
	Srv *auth.AuthService
}

type LoginRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (handler *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body LoginRequestBody

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	user, err := handler.Srv.AuthenticateWithPassword(ctx, auth.PasswordCredentials{
		Username: body.Username,
		Password: body.Password,
	})

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		return
	}

	session, err := handler.Srv.CreateSession(ctx, user)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	cookie := &http.Cookie{
		Name:     "sessionID",
		Value:    session.ID,
		Quoted:   false,
		Expires:  session.ExpiresAt,
		MaxAge:   int(session.ExpiresAt.Sub(time.Now()).Seconds()),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, cookie)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(session)
}

func (handler *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body RegisterRequestBody

	err := json.NewDecoder(r.Body).Decode(&body)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	_, err = handler.Srv.Register(ctx, auth.PasswordCredentials{
		Username: body.Username,
		Password: body.Password,
	})

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (handler *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	sessionCookie := &http.Cookie{
		Name:     "sessionID",
		Value:    "",
		Quoted:   false,
		Expires:  time.Time{},
		MaxAge:   0,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	// TODO: Revoke session

	http.SetCookie(w, sessionCookie)
	w.WriteHeader(http.StatusOK)
}
