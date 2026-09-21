package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/adolfooes/api_faker/internal/api/middleware"
	"github.com/gorilla/mux"
)

func stubOK(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestJWTMiddleware_NoAuthHeader_Returns401(t *testing.T) {
	r := mux.NewRouter()
	secured := r.PathPrefix("/api").Subrouter()
	secured.Use(middleware.JWTMiddleware)
	secured.HandleFunc("/scenario", stubOK).Methods("POST")

	req := httptest.NewRequest("POST", "/api/scenario", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWTMiddleware_InvalidToken_Returns401(t *testing.T) {
	os.Setenv("JWT_SECRET_KEY", "test-secret")

	r := mux.NewRouter()
	secured := r.PathPrefix("/api").Subrouter()
	secured.Use(middleware.JWTMiddleware)
	secured.HandleFunc("/project", stubOK).Methods("GET")

	req := httptest.NewRequest("GET", "/api/project", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestMockRoute_NoAuth_NotUnauthorized(t *testing.T) {
	r := mux.NewRouter()

	// Mock route is public — not under the secured subrouter (mirrors router.go)
	r.HandleFunc("/api/mock/{project_id}/{path:.*}", stubOK).Methods("GET", "POST", "PUT", "DELETE", "PATCH")

	secured := r.PathPrefix("/api").Subrouter()
	secured.Use(middleware.JWTMiddleware)
	secured.HandleFunc("/scenario", stubOK).Methods("POST")

	req := httptest.NewRequest("GET", "/api/mock/1/some/nested/path", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusUnauthorized {
		t.Error("mock endpoint must not require auth")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 from public mock route, got %d", w.Code)
	}
}
