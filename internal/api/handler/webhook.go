package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/adolfooes/api_faker/config"
	"github.com/adolfooes/api_faker/pkg/utils/crud"
	"github.com/adolfooes/api_faker/pkg/utils/response"
)

// WebhookDispatchInput is the payload for POST /api/webhook/dispatch.
type WebhookDispatchInput struct {
	ProjectID   *int64            `json:"project_id,omitempty"`
	TargetURL   string            `json:"target_url"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	ContentType string            `json:"content_type"`
	Body        interface{}       `json:"body"`
}

// WebhookDispatchOutput is returned by POST /api/webhook/dispatch.
type WebhookDispatchOutput struct {
	Status          int         `json:"status"`
	ResponseHeaders http.Header `json:"response_headers,omitempty"`
	ResponseBody    string      `json:"response_body,omitempty"`
	DurationMs      int64       `json:"duration_ms"`
	Error           string      `json:"error,omitempty"`
}

// currentOwnerID extracts and parses account_id injected into the request
// context by JWTMiddleware, matching the pattern used by every other handler
// (see project.go).
func currentOwnerID(r *http.Request) (int64, error) {
	ownerIDStr, ok := r.Context().Value(config.JWTAccountIDKey).(string)
	if !ok {
		return 0, fmt.Errorf("owner ID not found")
	}
	return strconv.ParseInt(ownerIDStr, 10, 64)
}

// WebhookDispatchHandler handles POST /api/webhook/dispatch: fires one outbound
// webhook on demand and returns the target's response.
func WebhookDispatchHandler(w http.ResponseWriter, r *http.Request) {
	// Corpo de entrada limitado: sem isso, um payload enorme vira armazenamento e
	// requisição de saída do mesmo tamanho.
	r.Body = http.MaxBytesReader(w, r.Body, maxWebhookResponseBytes)
	var input WebhookDispatchInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid request payload", err.Error(), nil, false)
		return
	}

	if input.TargetURL == "" {
		response.SendResponse(w, http.StatusBadRequest, "Validation failed", "target_url is required", nil, false)
		return
	}
	if err := validateTargetURL(input.TargetURL); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Validation failed", err.Error(), nil, false)
		return
	}

	ownerID, err := currentOwnerID(r)
	if err != nil {
		response.SendResponse(w, http.StatusUnauthorized, "Unauthorized: Owner ID not found", "", nil, false)
		return
	}
	if !allowDispatch(ownerID, time.Now()) {
		response.SendResponse(w, http.StatusTooManyRequests, "Too many webhook dispatches, slow down", "", nil, false)
		return
	}
	if input.ProjectID != nil {
		owned, err := validateProjectOwnership(int(*input.ProjectID), ownerID)
		if err != nil {
			response.SendResponse(w, http.StatusInternalServerError, "Failed to validate project ownership", err.Error(), nil, false)
			return
		}
		if !owned {
			response.SendResponse(w, http.StatusForbidden, "Unauthorized: project does not belong to this account", "", nil, false)
			return
		}
	}

	method := normalizeMethod(input.Method)
	contentType := normalizeContentType(input.ContentType)

	ctx := templateContext{}
	headers := interpolateHeaders(input.Headers, ctx)
	body := interpolateModel(input.Body, ctx)

	bodyBytes, err := encodeWebhookBody(contentType, body)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Failed to encode body", err.Error(), nil, false)
		return
	}

	outcome := dispatchWebhookRequest(input.TargetURL, method, contentType, headers, bodyBytes, config.GetWebhookTimeout())

	insertWebhookDispatchLog(input.ProjectID, ownerID, input.TargetURL, method, contentType, headers, string(bodyBytes), outcome)

	output := WebhookDispatchOutput{
		Status:          outcome.Status,
		ResponseHeaders: outcome.ResponseHeaders,
		ResponseBody:    outcome.ResponseBody,
		DurationMs:      outcome.DurationMs,
	}
	if outcome.Err != nil {
		output.Error = outcome.Err.Error()
	}
	response.SendResponse(w, http.StatusOK, "Webhook dispatched", "", output, false)
}

// WebhookDispatchRecord is one row of the webhook_dispatch log.
type WebhookDispatchRecord struct {
	ID              int64           `json:"id"`
	ProjectID       *int64          `json:"project_id,omitempty"`
	TargetURL       string          `json:"target_url"`
	Method          string          `json:"method"`
	ContentType     string          `json:"content_type"`
	RequestHeaders  json.RawMessage `json:"request_headers,omitempty"`
	RequestBody     string          `json:"request_body,omitempty"`
	ResponseStatus  *int            `json:"response_status,omitempty"`
	ResponseHeaders json.RawMessage `json:"response_headers,omitempty"`
	ResponseBody    string          `json:"response_body,omitempty"`
	Error           string          `json:"error,omitempty"`
	DurationMs      *int64          `json:"duration_ms,omitempty"`
	DispatchedAt    time.Time       `json:"dispatched_at"`
}

// ListWebhookDispatchesHandler handles GET /api/webhook/dispatch, optionally
// filtered by ?project_id=, listing the most recent dispatches belonging to the
// authenticated account only (any origin: admin-triggered or scenario-chained).
func ListWebhookDispatchesHandler(w http.ResponseWriter, r *http.Request) {
	const baseQuery = `
		SELECT id, project_id, target_url, method, content_type, request_headers, request_body,
		       response_status, response_headers, response_body, error, duration_ms, dispatched_at
		FROM webhook_dispatch`

	ownerID, err := currentOwnerID(r)
	if err != nil {
		response.SendResponse(w, http.StatusUnauthorized, "Unauthorized: Owner ID not found", "", nil, false)
		return
	}

	var rows []map[string]interface{}
	if projectIDStr := r.URL.Query().Get("project_id"); projectIDStr != "" {
		projectID, convErr := strconv.ParseInt(projectIDStr, 10, 64)
		if convErr != nil {
			response.SendResponse(w, http.StatusBadRequest, "Invalid project_id", convErr.Error(), nil, false)
			return
		}
		owned, ownErr := validateProjectOwnership(int(projectID), ownerID)
		if ownErr != nil {
			response.SendResponse(w, http.StatusInternalServerError, "Failed to validate project ownership", ownErr.Error(), nil, false)
			return
		}
		if !owned {
			response.SendResponse(w, http.StatusForbidden, "Unauthorized: project does not belong to this account", "", nil, false)
			return
		}
		rows, _, err = crud.Raw(baseQuery+` WHERE owner_id = $1 AND project_id = $2 ORDER BY dispatched_at DESC LIMIT 100`, ownerID, projectID)
	} else {
		rows, _, err = crud.Raw(baseQuery+` WHERE owner_id = $1 ORDER BY dispatched_at DESC LIMIT 100`, ownerID)
	}
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to fetch webhook dispatches", err.Error(), nil, false)
		return
	}

	records := make([]WebhookDispatchRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, webhookDispatchRecordFromRow(row))
	}
	response.SendResponse(w, http.StatusOK, "Webhook dispatches retrieved successfully", "", records, false)
}
