package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestEncodeFormBody_NestedTwoLevels(t *testing.T) {
	body := map[string]interface{}{
		"event": "invoice.status_changed",
		"data": map[string]interface{}{
			"id":     "abc123",
			"status": "paid",
			"payer": map[string]interface{}{
				"cpf": "12345678900",
			},
		},
	}
	got := encodeFormBody(body)
	want := "data[id]=abc123&data[payer][cpf]=12345678900&data[status]=paid&event=invoice.status_changed"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestEncodeWebhookBody_JSON(t *testing.T) {
	body := map[string]interface{}{"event": "invoice.created", "data": map[string]interface{}{"id": "1"}}
	got, err := encodeWebhookBody("application/json", body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("expected valid JSON, got %q: %v", got, err)
	}
	if decoded["event"] != "invoice.created" {
		t.Errorf("expected event field preserved, got %v", decoded["event"])
	}
}

func TestEncodeWebhookBody_RawStringPassthrough(t *testing.T) {
	raw := "event=invoice.created&data[id]=1"
	got, err := encodeWebhookBody("application/x-www-form-urlencoded", raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != raw {
		t.Errorf("expected raw string passthrough, got %q", got)
	}
}

func TestDispatchWebhookRequest_DeliversExactBodyAndHeaders_FormEncoded(t *testing.T) {
	var gotBody, gotContentType, gotCustomHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		gotContentType = r.Header.Get("Content-Type")
		gotCustomHeader = r.Header.Get("X-Iugu-Signature")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	bodyBytes, err := encodeWebhookBody("application/x-www-form-urlencoded", map[string]interface{}{
		"event": "invoice.status_changed",
		"data":  map[string]interface{}{"id": "inv_1", "status": "paid"},
	})
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	outcome := dispatchWebhookRequest(server.URL, "POST", "application/x-www-form-urlencoded",
		map[string]string{"X-Iugu-Signature": "sig-123"}, bodyBytes, 2*time.Second)

	if outcome.Err != nil {
		t.Fatalf("unexpected dispatch error: %v", outcome.Err)
	}
	if outcome.Status != http.StatusOK {
		t.Errorf("expected 200, got %d", outcome.Status)
	}
	wantBody := "data[id]=inv_1&data[status]=paid&event=invoice.status_changed"
	if gotBody != wantBody {
		t.Errorf("body mismatch: got %q, want %q", gotBody, wantBody)
	}
	if gotContentType != "application/x-www-form-urlencoded" {
		t.Errorf("expected form content-type, got %q", gotContentType)
	}
	if gotCustomHeader != "sig-123" {
		t.Errorf("expected custom header forwarded, got %q", gotCustomHeader)
	}
}

func TestDispatchWebhookRequest_DeliversExactBodyAndHeaders_JSON(t *testing.T) {
	var gotBody map[string]interface{}
	var gotContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	bodyBytes, err := encodeWebhookBody("application/json", map[string]interface{}{"event": "invoice.created"})
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	outcome := dispatchWebhookRequest(server.URL, "POST", "application/json", nil, bodyBytes, 2*time.Second)

	if outcome.Err != nil {
		t.Fatalf("unexpected dispatch error: %v", outcome.Err)
	}
	if outcome.Status != http.StatusCreated {
		t.Errorf("expected 201, got %d", outcome.Status)
	}
	if gotContentType != "application/json" {
		t.Errorf("expected json content-type, got %q", gotContentType)
	}
	if gotBody["event"] != "invoice.created" {
		t.Errorf("expected event field delivered, got %v", gotBody["event"])
	}
}

func TestDispatchWebhookRequest_Timeout_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	outcome := dispatchWebhookRequest(server.URL, "POST", "application/json", nil, []byte("{}"), 5*time.Millisecond)

	if outcome.Err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(outcome.Err.Error(), "deadline exceeded") && !strings.Contains(outcome.Err.Error(), "Client.Timeout") {
		t.Errorf("expected timeout-related error, got: %v", outcome.Err)
	}
}

func TestDispatchOneWebhook_TemplateResolvesRequestAndResponseFields(t *testing.T) {
	var gotBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	wh := ScenarioWebhookInput{
		TargetURL:   server.URL,
		ContentType: "application/json",
		Body: map[string]interface{}{
			"event": "invoice.status_changed",
			"data": map[string]interface{}{
				"id":        "{{response.data.id}}",
				"payer_cpf": "{{request.body.cpf}}",
				"status":    "paid",
			},
		},
	}
	ctx := templateContext{
		body:     map[string]interface{}{"cpf": "12345678900"},
		response: map[string]interface{}{"data": map[string]interface{}{"id": "inv_42"}},
	}

	_, _, outcome := dispatchOneWebhook(wh, ctx)
	if outcome.Err != nil {
		t.Fatalf("unexpected dispatch error: %v", outcome.Err)
	}

	data, ok := gotBody["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %v", gotBody["data"])
	}
	if data["id"] != "inv_42" {
		t.Errorf("expected response id resolved, got %v", data["id"])
	}
	if data["payer_cpf"] != "12345678900" {
		t.Errorf("expected request body field resolved, got %v", data["payer_cpf"])
	}
}

func TestParseScenarioWebhooks_NilOrInvalid_ReturnsEmpty(t *testing.T) {
	if got := parseScenarioWebhooks(nil); got != nil {
		t.Errorf("expected nil for nil input, got %v", got)
	}
	if got := parseScenarioWebhooks(123); got != nil {
		t.Errorf("expected nil for non-bytes/string input, got %v", got)
	}
	if got := parseScenarioWebhooks([]byte("not json")); got != nil {
		t.Errorf("expected nil for invalid JSON, got %v", got)
	}
}

func TestValidateTargetURL_BlocksLoopbackAndPrivateWhenNotAllowed(t *testing.T) {
	saved := webhookAllowedHosts
	webhookAllowedHosts = parseAllowedHosts("nothing.invalid")
	defer func() { webhookAllowedHosts = saved }()

	targets := []string{
		"http://127.0.0.1:8080/hook",
		"http://localhost/hook",
		"http://169.254.169.254/latest/meta-data/",
		"http://10.0.0.5/hook",
		"http://192.168.1.1/hook",
		"http://[::1]/hook",
	}
	for _, target := range targets {
		if err := validateTargetURL(target); err == nil {
			t.Errorf("expected %q to be rejected as SSRF target, got nil error", target)
		}
	}
}

func TestValidateTargetURL_AllowlistedHostIsAllowedButMetadataNeverIs(t *testing.T) {
	saved := webhookAllowedHosts
	webhookAllowedHosts = parseAllowedHosts("localhost,127.0.0.1,169.254.169.254")
	defer func() { webhookAllowedHosts = saved }()

	for _, target := range []string{"http://localhost:8080/hook", "http://127.0.0.1/hook"} {
		if err := validateTargetURL(target); err != nil {
			t.Errorf("expected allowlisted %q to be allowed, got: %v", target, err)
		}
	}
	if err := validateTargetURL("http://169.254.169.254/latest/meta-data/"); err == nil {
		t.Error("link-local metadata address must be blocked even when allowlisted")
	}
}

func TestSafeDialContext_BlocksNonAllowlistedLoopbackAtConnectTime(t *testing.T) {
	saved := webhookAllowedHosts
	webhookAllowedHosts = parseAllowedHosts("nothing.invalid")
	defer func() { webhookAllowedHosts = saved }()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer srv.Close()
	out := dispatchWebhookRequest(srv.URL, "POST", "application/json", nil, []byte("{}"), 2*time.Second)
	if out.Err == nil {
		t.Fatal("expected dial to non-allowlisted loopback to be blocked")
	}
}

func TestValidateTargetURL_RejectsNonHTTPScheme(t *testing.T) {
	if err := validateTargetURL("file:///etc/passwd"); err == nil {
		t.Error("expected non-http(s) scheme to be rejected")
	}
}

func TestValidateTargetURL_AllowsPublicHost(t *testing.T) {
	if err := validateTargetURL("https://example.com/webhook"); err != nil {
		t.Errorf("expected public host to be allowed, got: %v", err)
	}
}

func TestMaskSensitiveHeaders_RedactsKnownSensitiveNames(t *testing.T) {
	headers := map[string]string{
		"X-Iugu-Signature": "sig-123",
		"Authorization":    "Bearer secret-token",
		"X-Api-Key":        "abc",
		"Content-Type":     "application/json",
	}
	got := maskSensitiveHeaders(headers)
	for _, k := range []string{"X-Iugu-Signature", "Authorization", "X-Api-Key"} {
		if got[k] != maskedHeaderValue {
			t.Errorf("expected %q to be redacted, got %q", k, got[k])
		}
	}
	if got["Content-Type"] != "application/json" {
		t.Errorf("expected non-sensitive header preserved, got %q", got["Content-Type"])
	}
}

func TestParseScenarioWebhooks_ValidList(t *testing.T) {
	raw := []byte(`[{"target_url":"http://example.com","delay_ms":100}]`)
	got := parseScenarioWebhooks(raw)
	if len(got) != 1 {
		t.Fatalf("expected 1 webhook, got %d", len(got))
	}
	if got[0].TargetURL != "http://example.com" || got[0].DelayMs != 100 {
		t.Errorf("unexpected decoded webhook: %+v", got[0])
	}
}

func TestAllowDispatch_RateLimitsPerOwner(t *testing.T) {
	now := time.Now()
	owner := int64(987654321)
	allowed := 0
	for i := 0; i < 30; i++ {
		if allowDispatch(owner, now) {
			allowed++
		}
	}
	if allowed != 20 {
		t.Fatalf("expected burst of 20 dispatches at the same instant, got %d", allowed)
	}
	if !allowDispatch(owner, now.Add(time.Second)) {
		t.Fatal("expected tokens to refill after one second")
	}
	if !allowDispatch(owner+1, now) {
		t.Fatal("expected a different owner to have its own bucket")
	}
}
