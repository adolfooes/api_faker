package handler

import (
	"testing"
)

func TestGenerateUUID_UniqueAcrossCalls(t *testing.T) {
	a := generateUUID()
	b := generateUUID()
	if a == b {
		t.Errorf("expected different UUIDs, got same: %s", a)
	}
}

func TestResolveToken_RequestBodyField_ReturnsValue(t *testing.T) {
	ctx := templateContext{
		body: map[string]interface{}{"name": "alice"},
	}
	got := resolveToken("request.body.name", ctx)
	if got != "alice" {
		t.Errorf("expected alice, got %v", got)
	}
}

func TestResolveToken_NestedBodyField(t *testing.T) {
	ctx := templateContext{
		body: map[string]interface{}{
			"user": map[string]interface{}{"age": float64(30)},
		},
	}
	got := resolveToken("request.body.user.age", ctx)
	if got != float64(30) {
		t.Errorf("expected 30.0, got %v (%T)", got, got)
	}
}

func TestResolveToken_UnknownToken_RemainsLiteral(t *testing.T) {
	ctx := templateContext{}
	got := resolveToken("totally_unknown", ctx)
	want := "{{totally_unknown}}"
	if got != want {
		t.Errorf("expected %q, got %v", want, got)
	}
}

func TestInterpolateModel_NumberTypePreserved(t *testing.T) {
	ctx := templateContext{}
	input := map[string]interface{}{"count": float64(42)}
	result := interpolateModel(input, ctx)
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if m["count"] != float64(42) {
		t.Errorf("number type not preserved: got %v (%T)", m["count"], m["count"])
	}
}

func TestInterpolateModel_BoolTypePreserved(t *testing.T) {
	ctx := templateContext{}
	input := map[string]interface{}{"active": true}
	result := interpolateModel(input, ctx)
	m, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", result)
	}
	if m["active"] != true {
		t.Errorf("bool type not preserved: got %v (%T)", m["active"], m["active"])
	}
}

func TestInterpolateString_SingleToken_ReturnsTypedValue(t *testing.T) {
	ctx := templateContext{
		body: map[string]interface{}{"score": float64(99)},
	}
	result := interpolateString("{{request.body.score}}", ctx)
	if result != float64(99) {
		t.Errorf("expected float64(99) for single token, got %v (%T)", result, result)
	}
}

func TestInterpolateString_EmbeddedToken_ReturnsString(t *testing.T) {
	ctx := templateContext{
		body: map[string]interface{}{"name": "world"},
	}
	result := interpolateString("hello {{request.body.name}}!", ctx)
	want := "hello world!"
	if result != want {
		t.Errorf("expected %q, got %v", want, result)
	}
}
