package handler

import (
	"encoding/json"
	"log"
	"time"

	"github.com/adolfooes/api_faker/pkg/utils/crud"
)

// webhookDispatchRecordFromRow converts one crud.Raw result row (column name to
// driver-scanned value) into a WebhookDispatchRecord, tolerating the NULL columns.
func webhookDispatchRecordFromRow(row map[string]interface{}) WebhookDispatchRecord {
	rec := WebhookDispatchRecord{
		TargetURL:    dispatchRowString(row["target_url"]),
		Method:       dispatchRowString(row["method"]),
		ContentType:  dispatchRowString(row["content_type"]),
		RequestBody:  dispatchRowString(row["request_body"]),
		ResponseBody: dispatchRowString(row["response_body"]),
		Error:        dispatchRowString(row["error"]),
	}
	if id, ok := row["id"].(int64); ok {
		rec.ID = id
	}
	if projectID, ok := row["project_id"].(int64); ok {
		rec.ProjectID = &projectID
	}
	if raw := dispatchRowJSON(row["request_headers"]); raw != nil {
		rec.RequestHeaders = raw
	}
	if status, ok := row["response_status"].(int64); ok {
		v := int(status)
		rec.ResponseStatus = &v
	}
	if raw := dispatchRowJSON(row["response_headers"]); raw != nil {
		rec.ResponseHeaders = raw
	}
	if durationMs, ok := row["duration_ms"].(int64); ok {
		rec.DurationMs = &durationMs
	}
	if dispatchedAt, ok := row["dispatched_at"].(time.Time); ok {
		rec.DispatchedAt = dispatchedAt
	}
	return rec
}

// dispatchRowString reads a possibly-NULL text column, tolerating both string
// and []byte scan results depending on the driver.
func dispatchRowString(v interface{}) string {
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	default:
		return ""
	}
}

// dispatchRowJSON reads a possibly-NULL jsonb column, tolerating both string
// and []byte scan results depending on the driver.
func dispatchRowJSON(v interface{}) json.RawMessage {
	switch b := v.(type) {
	case []byte:
		return json.RawMessage(b)
	case string:
		return json.RawMessage(b)
	default:
		return nil
	}
}

// insertWebhookDispatchLog persists one dispatch attempt for later inspection via
// GET /api/webhook/dispatch, scoped to ownerID so ListWebhookDispatchesHandler
// can never return another account's data. Sensitive header values (tokens,
// signatures) are redacted before persisting. Logging failures never fail the
// caller's request.
func insertWebhookDispatchLog(projectID *int64, ownerID int64, targetURL, method, contentType string, headers map[string]string, sentBody string, outcome dispatchOutcome) {
	headersJSON, _ := json.Marshal(maskSensitiveHeaders(headers))

	var pid interface{}
	if projectID != nil {
		pid = *projectID
	}
	var respStatus interface{}
	if outcome.Status != 0 {
		respStatus = outcome.Status
	}
	var respHeadersJSON interface{}
	if outcome.ResponseHeaders != nil {
		b, _ := json.Marshal(maskSensitiveHTTPHeaders(outcome.ResponseHeaders))
		respHeadersJSON = string(b)
	}
	var errText interface{}
	if outcome.Err != nil {
		errText = outcome.Err.Error()
	}

	columns := []string{
		"owner_id", "project_id", "target_url", "method", "content_type", "request_headers", "request_body",
		"response_status", "response_headers", "response_body", "error", "duration_ms",
	}
	values := []interface{}{
		ownerID, pid, targetURL, method, contentType, string(headersJSON), sentBody,
		respStatus, respHeadersJSON, outcome.ResponseBody, errText, outcome.DurationMs,
	}
	if _, err := crud.Create("webhook_dispatch", columns, values); err != nil {
		log.Printf("webhook_dispatch: failed to persist dispatch log: %v", err)
	}
}
