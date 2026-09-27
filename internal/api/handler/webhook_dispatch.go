package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/adolfooes/api_faker/config"
)

// ScenarioWebhookInput is a webhook chained to a scenario endpoint, fired
// asynchronously after the mocked response has been served.
type ScenarioWebhookInput struct {
	DelayMs     int               `json:"delay_ms"`
	TargetURL   string            `json:"target_url"`
	Method      string            `json:"method"`
	ContentType string            `json:"content_type"`
	Headers     map[string]string `json:"headers"`
	Body        interface{}       `json:"body"`
}

// dispatchOutcome is the result of one outbound HTTP call, with no DB access,
// so it stays unit-testable on its own.
type dispatchOutcome struct {
	Status          int
	ResponseHeaders http.Header
	ResponseBody    string
	DurationMs      int64
	Err             error

	// ResolvedMethod and ResolvedContentType carry the already-defaulted
	// method/content_type out of dispatchOneWebhook, so callers logging the
	// dispatch never recompute the same default.
	ResolvedMethod      string
	ResolvedContentType string
}

// normalizeMethod upper-cases method, defaulting an empty value to POST.
func normalizeMethod(method string) string {
	m := strings.ToUpper(strings.TrimSpace(method))
	if m == "" {
		return http.MethodPost
	}
	return m
}

// normalizeContentType defaults an empty content_type to application/json.
func normalizeContentType(contentType string) string {
	if contentType == "" {
		return "application/json"
	}
	return contentType
}

// dispatchWebhookRequest performs the outbound HTTP call. No retry on failure or
// timeout — the caller decides what to do with the error.
// maxWebhookResponseBytes limita o corpo da resposta do alvo guardado e devolvido.
const maxWebhookResponseBytes = 1 << 20

func dispatchWebhookRequest(targetURL, method, contentType string, headers map[string]string, bodyBytes []byte, timeout time.Duration) dispatchOutcome {
	start := time.Now()

	httpReq, err := http.NewRequest(method, targetURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return dispatchOutcome{Err: err, DurationMs: time.Since(start).Milliseconds()}
	}
	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	// No redirect following: a target could otherwise 30x-redirect the request
	// to an internal address after passing validateTargetURL (SSRF via redirect).
	client := &http.Client{
		Timeout:   timeout,
		Transport: &http.Transport{DialContext: safeDialContext, Proxy: nil},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(httpReq)
	duration := time.Since(start).Milliseconds()
	if err != nil {
		return dispatchOutcome{Err: err, DurationMs: duration}
	}
	defer resp.Body.Close()

	// Resposta do alvo limitada: um host lento ou malicioso não esgota a memória.
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, maxWebhookResponseBytes))
	return dispatchOutcome{
		Status:          resp.StatusCode,
		ResponseHeaders: resp.Header,
		ResponseBody:    string(respBody),
		DurationMs:      duration,
	}
}

// dispatchOneWebhook resolves templates, encodes and sends a single scenario
// webhook. No DB access, so it can be unit-tested without a database.
func dispatchOneWebhook(wh ScenarioWebhookInput, ctx templateContext) (headers map[string]string, sentBody string, outcome dispatchOutcome) {
	contentType := normalizeContentType(wh.ContentType)
	method := normalizeMethod(wh.Method)

	headers = interpolateHeaders(wh.Headers, ctx)
	body := interpolateModel(wh.Body, ctx)

	bodyBytes, err := encodeWebhookBody(contentType, body)
	if err != nil {
		return headers, "", dispatchOutcome{Err: err, ResolvedMethod: method, ResolvedContentType: contentType}
	}

	outcome = dispatchWebhookRequest(wh.TargetURL, method, contentType, headers, bodyBytes, config.GetWebhookTimeout())
	outcome.ResolvedMethod = method
	outcome.ResolvedContentType = contentType
	return headers, string(bodyBytes), outcome
}

// fireScenarioWebhooks dispatches every webhook chained to a scenario endpoint,
// each after its own delay_ms, asynchronously so the mocked response the caller
// already received is never held up. Every dispatch, success or failure, is
// logged to webhook_dispatch, attributed to the scenario's project owner.
func fireScenarioWebhooks(webhooks []ScenarioWebhookInput, projectID *int64, ownerID int64, ctx templateContext) {
	for _, wh := range webhooks {
		wh := wh
		go func() {
			if wh.DelayMs > 0 {
				time.Sleep(time.Duration(wh.DelayMs) * time.Millisecond)
			}
			method := normalizeMethod(wh.Method)
			contentType := normalizeContentType(wh.ContentType)
			if err := validateTargetURL(wh.TargetURL); err != nil {
				insertWebhookDispatchLog(projectID, ownerID, wh.TargetURL, method, contentType, nil, "", dispatchOutcome{Err: err})
				return
			}
			headers, sentBody, outcome := dispatchOneWebhook(wh, ctx)
			insertWebhookDispatchLog(projectID, ownerID, wh.TargetURL, outcome.ResolvedMethod, outcome.ResolvedContentType, headers, sentBody, outcome)
		}()
	}
}

// parseScenarioWebhooks decodes the url_config.webhooks JSONB column. Absent or
// invalid data yields no webhooks, keeping scenarios without this field unaffected.
func parseScenarioWebhooks(raw interface{}) []ScenarioWebhookInput {
	var rawBytes []byte
	switch v := raw.(type) {
	case []byte:
		rawBytes = v
	case string:
		rawBytes = []byte(v)
	default:
		return nil
	}
	var webhooks []ScenarioWebhookInput
	if err := json.Unmarshal(rawBytes, &webhooks); err != nil {
		return nil
	}
	return webhooks
}
