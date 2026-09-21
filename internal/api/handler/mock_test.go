package handler

import (
	"testing"
)

func TestMatchPath_SingleParam_Matches(t *testing.T) {
	ok, params := matchPath("/a/{id}/b", "/a/123/b")
	if !ok {
		t.Fatal("expected match")
	}
	if params["id"] != "123" {
		t.Errorf("expected id=123, got %q", params["id"])
	}
}

func TestMatchPath_ExtraSegment_NoMatch(t *testing.T) {
	ok, _ := matchPath("/a/{id}/b", "/a/123/456/b")
	if ok {
		t.Error("expected no match when segment count differs")
	}
}

func TestMatchPath_LiteralMismatch_NoMatch(t *testing.T) {
	ok, _ := matchPath("/a/b/c", "/a/x/c")
	if ok {
		t.Error("expected no match when literal segment differs")
	}
}

func TestFindMatchingURLConfig_LiteralBeatsParam_LiteralFirst(t *testing.T) {
	candidates := []map[string]interface{}{
		{"path": "/users/me"},
		{"path": "/users/{id}"},
	}
	cfg, _ := findMatchingURLConfig(candidates, "/users/me")
	if cfg == nil {
		t.Fatal("expected a match")
	}
	if cfg["path"] != "/users/me" {
		t.Errorf("literal should win, got %v", cfg["path"])
	}
}

func TestFindMatchingURLConfig_LiteralBeatsParam_ParamFirst(t *testing.T) {
	// param listed before literal in slice — literal must still win
	candidates := []map[string]interface{}{
		{"path": "/users/{id}"},
		{"path": "/users/me"},
	}
	cfg, _ := findMatchingURLConfig(candidates, "/users/me")
	if cfg == nil {
		t.Fatal("expected a match")
	}
	if cfg["path"] != "/users/me" {
		t.Errorf("literal should win regardless of order, got %v", cfg["path"])
	}
}

func TestFindMatchingURLConfig_ParamFallback(t *testing.T) {
	candidates := []map[string]interface{}{
		{"path": "/users/{id}"},
	}
	cfg, params := findMatchingURLConfig(candidates, "/users/42")
	if cfg == nil {
		t.Fatal("expected param match")
	}
	if params["id"] != "42" {
		t.Errorf("expected id=42, got %q", params["id"])
	}
}

func TestFindMatchingURLConfig_NoMatchDifferentPrefix(t *testing.T) {
	candidates := []map[string]interface{}{
		{"path": "/users/{id}"},
	}
	cfg, _ := findMatchingURLConfig(candidates, "/orders/42")
	if cfg != nil {
		t.Error("expected no match for different prefix")
	}
}
