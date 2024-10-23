package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dpbrackin/ready-set-go/router"
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("TEST"))
}

func TestRouteFunc(t *testing.T) {
	router := router.NewRootRouter()
	router.RouteFunc("GET /", testHandler)
	mux := router.Mux()

	req := httptest.NewRequest("GET", "/", nil)

	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}
}

func TestRouteGroup(t *testing.T) {
	router := router.NewRootRouter()

	g := router.Group("/group")
	g.RouteFunc("/test", testHandler)

	mux := router.Mux()

	req := httptest.NewRequest("GET", "/group/test", nil)

	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}
}
