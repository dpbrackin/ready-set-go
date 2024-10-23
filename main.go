package main

import (
	"net/http"

	"github.com/dpbrackin/ready-set-go/router"
)

func main() {
	root := router.NewRootRouter()
	root.Use(LoggingMiddleware)

	unauthenticatedGroup := root.Group("")
	unauthenticatedGroup.RouteFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {})

	http.ListenAndServe(":3000", root.Mux())
}
