package handler_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/adolfooes/api_faker/config"
	"github.com/adolfooes/api_faker/internal/api/handler"
	internaldb "github.com/adolfooes/api_faker/internal/db"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

var initDBOnce sync.Once

func initTestDB(t *testing.T) {
	t.Helper()
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	initDBOnce.Do(func() {
		internaldb.InitDB(dbURL)
		applyMigrations(t)
	})
}

// applyMigrations monta o schema no banco de teste.
//
// internaldb.RunMigrations nao serve aqui: ele aponta para file:///migrations,
// caminho absoluto que so existe dentro do container de producao. E um teste de
// integracao que depende de alguem ter criado as tabelas antes falha com
// "relation does not exist" -- que foi exatamente o que aconteceu.
//
// Idempotente: erro de objeto ja existente e ignorado, para a suite poder rodar
// varias vezes contra o mesmo banco.
func applyMigrations(t *testing.T) {
	t.Helper()
	dir := filepath.Join("..", "..", "db", "migrations")
	files, err := filepath.Glob(filepath.Join(dir, "*.up.sql"))
	if err != nil || len(files) == 0 {
		t.Fatalf("migrations nao encontradas em %s: %v", dir, err)
	}
	sort.Strings(files)
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("ler %s: %v", f, err)
		}
		if _, err := internaldb.GetDB().Exec(string(content)); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				t.Fatalf("aplicar %s: %v", filepath.Base(f), err)
			}
		}
	}
}

func createTestAccount(t *testing.T, db *sql.DB) (int64, func()) {
	t.Helper()
	hashed, err := bcrypt.GenerateFromPassword([]byte("TestPass1!"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	email := fmt.Sprintf("test_%d@example.com", time.Now().UnixNano())
	var id int64
	err = db.QueryRow(
		`INSERT INTO account (email, password) VALUES ($1, $2) RETURNING id`,
		email, string(hashed),
	).Scan(&id)
	if err != nil {
		t.Fatalf("create test account: %v", err)
	}
	return id, func() {
		db.Exec(`DELETE FROM account WHERE id = $1`, id)
	}
}

func buildTestRouter(accountID int64) *mux.Router {
	r := mux.NewRouter()
	secured := r.PathPrefix("/api").Subrouter()
	secured.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), config.JWTAccountIDKey, strconv.FormatInt(accountID, 10))
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
	secured.HandleFunc("/scenario", handler.CreateScenarioHandler).Methods("POST")
	secured.HandleFunc("/scenario/{project_id:[0-9]+}", handler.GetScenarioHandler).Methods("GET")
	return r
}

func makeScenarioBody(projectName string, percentage int) string {
	return fmt.Sprintf(`{
		"project": {"name": %q, "description": "test"},
		"endpoints": [{
			"path": "/test",
			"method": "GET",
			"description": "desc",
			"statuses": [{
				"http_status": 200,
				"percentage": %d,
				"model": {"ok": true}
			}]
		}]
	}`, projectName, percentage)
}

func TestCreateScenarioHandler_Idempotent(t *testing.T) {
	initTestDB(t)
	db := internaldb.GetDB()
	accountID, cleanup := createTestAccount(t, db)
	defer cleanup()

	r := buildTestRouter(accountID)
	projectName := fmt.Sprintf("idempotent-proj-%d", time.Now().UnixNano())
	body := makeScenarioBody(projectName, 100)

	// First call
	req1 := httptest.NewRequest("POST", "/api/scenario", strings.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("first POST expected 200, got %d: %s", w1.Code, w1.Body.String())
	}

	var resp1 struct {
		Data struct {
			ProjectID int64 `json:"project_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w1.Body).Decode(&resp1); err != nil {
		t.Fatalf("decode first response: %v", err)
	}

	// Second call — same payload
	req2 := httptest.NewRequest("POST", "/api/scenario", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("second POST expected 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var resp2 struct {
		Data struct {
			ProjectID int64 `json:"project_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w2.Body).Decode(&resp2); err != nil {
		t.Fatalf("decode second response: %v", err)
	}

	if resp1.Data.ProjectID == 0 {
		t.Fatal("first response has no project_id")
	}
	if resp1.Data.ProjectID != resp2.Data.ProjectID {
		t.Errorf("idempotent call returned different project_id: %d vs %d",
			resp1.Data.ProjectID, resp2.Data.ProjectID)
	}

	// Verify exactly one project row with this name
	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM project WHERE name = $1 AND owner_id = $2`,
		projectName, accountID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("count query: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 project row, got %d", count)
	}
}

func TestCreateScenarioHandler_RollbackOnPartialError(t *testing.T) {
	initTestDB(t)
	db := internaldb.GetDB()
	accountID, cleanup := createTestAccount(t, db)
	defer cleanup()

	r := buildTestRouter(accountID)
	projectName := fmt.Sprintf("rollback-proj-%d", time.Now().UnixNano())

	// percentage=-1 passes Go validation (total=49≤100) but violates the
	// DB CHECK constraint (percentage BETWEEN 0 AND 100), triggering rollback.
	body := fmt.Sprintf(`{
		"project": {"name": %q},
		"endpoints": [{
			"path": "/test",
			"method": "GET",
			"statuses": [
				{"http_status": 200, "percentage": 50, "model": {}},
				{"http_status": 404, "percentage": -1, "model": {}}
			]
		}]
	}`, projectName)

	req := httptest.NewRequest("POST", "/api/scenario", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Fatal("expected non-200 response due to DB constraint violation")
	}

	// Verify no project was committed (transaction rolled back)
	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM project WHERE name = $1 AND owner_id = $2`,
		projectName, accountID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("count query: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 projects after rollback, got %d", count)
	}
}

func TestGetScenarioHandler_ReturnsCreatedData(t *testing.T) {
	initTestDB(t)
	db := internaldb.GetDB()
	accountID, cleanup := createTestAccount(t, db)
	defer cleanup()

	r := buildTestRouter(accountID)
	projectName := fmt.Sprintf("get-scenario-proj-%d", time.Now().UnixNano())

	// Create scenario
	createBody := fmt.Sprintf(`{
		"project": {"name": %q, "description": "get-test"},
		"endpoints": [{
			"path": "/items",
			"method": "GET",
			"description": "list items",
			"statuses": [{
				"http_status": 200,
				"percentage": 100,
				"model": {"items": []}
			}]
		}]
	}`, projectName)

	req := httptest.NewRequest("POST", "/api/scenario", strings.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("POST /api/scenario expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var createResp struct {
		Data struct {
			ProjectID int64 `json:"project_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(w.Body).Decode(&createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createResp.Data.ProjectID == 0 {
		t.Fatal("no project_id in create response")
	}

	// Fetch scenario
	getReq := httptest.NewRequest("GET",
		fmt.Sprintf("/api/scenario/%d", createResp.Data.ProjectID), nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("GET /api/scenario/%d expected 200, got %d: %s",
			createResp.Data.ProjectID, getW.Code, getW.Body.String())
	}

	var getResp struct {
		Data struct {
			Project struct {
				Name string `json:"name"`
			} `json:"project"`
			Endpoints []struct {
				Path    string `json:"path"`
				Method  string `json:"method"`
				Statuses []struct {
					HTTPStatus int `json:"http_status"`
					Percentage int `json:"percentage"`
				} `json:"statuses"`
			} `json:"endpoints"`
		} `json:"data"`
	}
	if err := json.NewDecoder(getW.Body).Decode(&getResp); err != nil {
		t.Fatalf("decode get response: %v", err)
	}

	if getResp.Data.Project.Name != projectName {
		t.Errorf("expected project name %q, got %q", projectName, getResp.Data.Project.Name)
	}
	if len(getResp.Data.Endpoints) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(getResp.Data.Endpoints))
	}
	ep := getResp.Data.Endpoints[0]
	if ep.Path != "/items" {
		t.Errorf("expected path /items, got %q", ep.Path)
	}
	if ep.Method != "GET" {
		t.Errorf("expected method GET, got %q", ep.Method)
	}
	if len(ep.Statuses) != 1 {
		t.Fatalf("expected 1 status, got %d", len(ep.Statuses))
	}
	if ep.Statuses[0].HTTPStatus != 200 {
		t.Errorf("expected http_status 200, got %d", ep.Statuses[0].HTTPStatus)
	}
	if ep.Statuses[0].Percentage != 100 {
		t.Errorf("expected percentage 100, got %d", ep.Statuses[0].Percentage)
	}

	// Verify DB state directly — project should exist
	var dbCount int
	db.QueryRow(
		`SELECT COUNT(*) FROM project WHERE id = $1 AND name = $2`,
		createResp.Data.ProjectID, projectName,
	).Scan(&dbCount)
	if dbCount != 1 {
		t.Errorf("expected 1 project in DB, got %d", dbCount)
	}
}
