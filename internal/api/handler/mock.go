package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/adolfooes/api_faker/pkg/utils/crud"
	"github.com/adolfooes/api_faker/pkg/utils/response"
	"github.com/gorilla/mux"
)

func validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path is required")
	}

	// Allows alphanumeric, slashes, dashes, underscores, and {param} segments
	regex := regexp.MustCompile(`^\/[a-zA-Z0-9\/\-_\{\}]*$`)

	if !regex.MatchString(path) {
		return fmt.Errorf("invalid path: path can only contain alphanumeric characters, slashes (/), dashes (-), underscores (_), and path parameters ({name})")
	}

	return nil
}

func validateProjectID(projectIDStr string) (int, error) {
	if projectIDStr == "" {
		return 0, fmt.Errorf("missing project ID in URL")
	}

	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid project ID")
	}

	return projectID, nil
}

// matchPath checks if requestPath matches pattern (supports {name} segments).
// Returns (matched, extracted params). Literal segments must match exactly;
// {name} captures exactly one segment. Segment count must be equal.
func matchPath(pattern, requestPath string) (bool, map[string]string) {
	patternSegs := strings.Split(strings.Trim(pattern, "/"), "/")
	requestSegs := strings.Split(strings.Trim(requestPath, "/"), "/")

	if len(patternSegs) != len(requestSegs) {
		return false, nil
	}

	params := make(map[string]string)
	for i, pSeg := range patternSegs {
		if strings.HasPrefix(pSeg, "{") && strings.HasSuffix(pSeg, "}") {
			name := pSeg[1 : len(pSeg)-1]
			params[name] = requestSegs[i]
		} else if pSeg != requestSegs[i] {
			return false, nil
		}
	}
	return true, params
}

// findMatchingURLConfig returns the best match from candidates for requestPath.
// Literal match (no {}) wins over pattern match regardless of order.
func findMatchingURLConfig(candidates []map[string]interface{}, requestPath string) (map[string]interface{}, map[string]string) {
	var patternMatch map[string]interface{}
	var patternParams map[string]string

	for _, candidate := range candidates {
		pattern, _ := candidate["path"].(string)
		matched, params := matchPath(pattern, requestPath)
		if !matched {
			continue
		}
		if !strings.Contains(pattern, "{") {
			return candidate, params
		}
		if patternMatch == nil {
			patternMatch = candidate
			patternParams = params
		}
	}
	return patternMatch, patternParams
}

func MockHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody map[string]interface{}
	if r.Body != nil {
		bodyBytes, _ := io.ReadAll(r.Body)
		r.Body.Close()
		_ = json.Unmarshal(bodyBytes, &requestBody)
	}

	vars := mux.Vars(r)
	path := "/" + vars["path"]

	if err := validatePath(path); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid path", err.Error(), nil, false)
		return
	}

	projectIDStr := vars["project_id"]

	projectID, err := validateProjectID(projectIDStr)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid project ID", err.Error(), nil, false)
		return
	}

	method := strings.ToUpper(r.Method)

	// Fetch all url_configs for project+method, then filter in memory for pattern matching
	candidates, err := crud.List("url_config", map[string]interface{}{
		"method":     method,
		"project_id": projectID,
	})
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to fetch URL configs", err.Error(), nil, false)
		return
	}

	urlConfig, pathParams := findMatchingURLConfig(candidates, path)
	if urlConfig == nil {
		response.SendResponse(w, http.StatusNotFound, "URL not configured for mocking", "", nil, false)
		return
	}

	// Fetch all the HTTP statuses and their percentages from url_http_status for this url_config
	httpStatuses, err := crud.List("url_http_status", map[string]interface{}{
		"url_id": urlConfig["id"],
	})
	if err != nil || len(httpStatuses) == 0 {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to fetch HTTP statuses", err.Error(), nil, false)
		return
	}

	// Randomize the response based on percentage
	selectedStatus := randomizeHTTPStatus(httpStatuses)

	// Fetch the corresponding response model from the response_model table
	responseModels, err := crud.List("response_model", map[string]interface{}{
		"url_http_status_id": selectedStatus["id"],
	})
	if err != nil || len(responseModels) == 0 {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to fetch response model", err.Error(), nil, false)
		return
	}
	responseModel := responseModels[0]

	ctx := templateContext{body: requestBody, pathParams: pathParams}
	var modelData interface{}
	modelParsed := false
	switch m := responseModel["model"].(type) {
	case string:
		if json.Unmarshal([]byte(m), &modelData) == nil {
			modelParsed = true
		}
	case []byte:
		if json.Unmarshal(m, &modelData) == nil {
			modelParsed = true
		}
	}

	var finalModel interface{}
	if modelParsed {
		finalModel = interpolateModel(modelData, ctx)
	} else {
		finalModel = responseModel["model"]
	}

	response.SendResponse(w, int(selectedStatus["http_status"].(int64)), "", "", finalModel, true)
}

// randomizeHTTPStatus selects a status based on the percentage distribution
func randomizeHTTPStatus(statuses []map[string]interface{}) map[string]interface{} {
	totalPercentage := 0
	for _, status := range statuses {
		totalPercentage += int(status["percentage"].(int64))
	}

	randomNumber := rand.Intn(100) // Random number between 0 and 99
	currentPercentage := 0

	for _, status := range statuses {
		currentPercentage += int(status["percentage"].(int64))
		if randomNumber < currentPercentage {
			return status
		}
	}

	// Default to the first one if something goes wrong
	return statuses[0]
}
