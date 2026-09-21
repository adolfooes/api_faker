package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/adolfooes/api_faker/config"
	internaldb "github.com/adolfooes/api_faker/internal/db"
	"github.com/adolfooes/api_faker/pkg/utils/crud"
	"github.com/adolfooes/api_faker/pkg/utils/response"
	"github.com/gorilla/mux"
)

type ScenarioStatusInput struct {
	HTTPStatus int             `json:"http_status"`
	Percentage int             `json:"percentage"`
	Model      json.RawMessage `json:"model"`
}

type ScenarioEndpointInput struct {
	Path        string                `json:"path"`
	Method      string                `json:"method"`
	Description string                `json:"description"`
	Statuses    []ScenarioStatusInput `json:"statuses"`
}

type ScenarioProjectInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ScenarioInput struct {
	Project   ScenarioProjectInput    `json:"project"`
	Endpoints []ScenarioEndpointInput `json:"endpoints"`
}

type ScenarioStatusResult struct {
	URLHTTPStatusID int64 `json:"url_http_status_id"`
	ResponseModelID int64 `json:"response_model_id"`
}

type ScenarioEndpointResult struct {
	URLConfigID int64                  `json:"url_config_id"`
	Statuses    []ScenarioStatusResult `json:"statuses"`
}

type ScenarioResult struct {
	ProjectID int64                    `json:"project_id"`
	Endpoints []ScenarioEndpointResult `json:"endpoints"`
}

func validateScenarioInput(input ScenarioInput) error {
	if input.Project.Name == "" {
		return fmt.Errorf("project.name is required")
	}
	for i, ep := range input.Endpoints {
		if ep.Path == "" {
			return fmt.Errorf("endpoints[%d].path is required", i)
		}
		if ep.Method == "" {
			return fmt.Errorf("endpoints[%d].method is required", i)
		}
		if len(ep.Statuses) == 0 {
			return fmt.Errorf("endpoints[%d].statuses must not be empty", i)
		}
		total := 0
		for _, s := range ep.Statuses {
			total += s.Percentage
		}
		if total > 100 {
			return fmt.Errorf("endpoints[%d]: sum of percentage (%d) exceeds 100", i, total)
		}
	}
	return nil
}

// CreateScenarioHandler handles POST /api/scenario.
// Creates or updates project, url_config, url_http_status and response_model in a single transaction.
// Idempotent by (project.name+owner_id, path, method).
func CreateScenarioHandler(w http.ResponseWriter, r *http.Request) {
	var input ScenarioInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid request payload", err.Error(), nil, false)
		return
	}

	if err := validateScenarioInput(input); err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Validation failed", err.Error(), nil, false)
		return
	}

	ownerIDStr, ok := r.Context().Value(config.JWTAccountIDKey).(string)
	if !ok {
		response.SendResponse(w, http.StatusUnauthorized, "Unauthorized: Owner ID not found", "", nil, false)
		return
	}
	ownerID, err := strconv.ParseInt(ownerIDStr, 10, 64)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid Owner ID format", "", nil, false)
		return
	}

	tx, err := internaldb.GetDB().Begin()
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to start transaction", err.Error(), nil, false)
		return
	}
	defer tx.Rollback() // no-op after Commit

	// Upsert project by (name, owner_id)
	var projectID int64
	err = tx.QueryRow(
		`SELECT id FROM project WHERE name = $1 AND owner_id = $2 LIMIT 1`,
		input.Project.Name, ownerID,
	).Scan(&projectID)
	if err == sql.ErrNoRows {
		err = tx.QueryRow(
			`INSERT INTO project (name, description, owner_id) VALUES ($1, $2, $3) RETURNING id`,
			input.Project.Name, input.Project.Description, ownerID,
		).Scan(&projectID)
		if err != nil {
			response.SendResponse(w, http.StatusInternalServerError, "Failed to create project", err.Error(), nil, false)
			return
		}
	} else if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to query project", err.Error(), nil, false)
		return
	} else {
		if _, err = tx.Exec(
			`UPDATE project SET description = $1 WHERE id = $2`,
			input.Project.Description, projectID,
		); err != nil {
			response.SendResponse(w, http.StatusInternalServerError, "Failed to update project", err.Error(), nil, false)
			return
		}
	}

	var endpointResults []ScenarioEndpointResult

	for _, ep := range input.Endpoints {
		// Upsert url_config by (project_id, path, method)
		var urlConfigID int64
		err = tx.QueryRow(
			`SELECT id FROM url_config WHERE project_id = $1 AND path = $2 AND method = $3 LIMIT 1`,
			projectID, ep.Path, ep.Method,
		).Scan(&urlConfigID)
		if err == sql.ErrNoRows {
			err = tx.QueryRow(
				`INSERT INTO url_config (path, method, description, project_id) VALUES ($1, $2, $3, $4) RETURNING id`,
				ep.Path, ep.Method, ep.Description, projectID,
			).Scan(&urlConfigID)
			if err != nil {
				response.SendResponse(w, http.StatusInternalServerError, "Failed to create url_config", err.Error(), nil, false)
				return
			}
		} else if err != nil {
			response.SendResponse(w, http.StatusInternalServerError, "Failed to query url_config", err.Error(), nil, false)
			return
		} else {
			if _, err = tx.Exec(
				`UPDATE url_config SET description = $1 WHERE id = $2`,
				ep.Description, urlConfigID,
			); err != nil {
				response.SendResponse(w, http.StatusInternalServerError, "Failed to update url_config", err.Error(), nil, false)
				return
			}
		}

		// O cenario declara o estado COMPLETO do endpoint: status que estava
		// cadastrado e nao vem no payload tem de sair. Sem isto o upsert por
		// (url_id, http_status) so acrescenta -- reaplicar o cenario com 422
		// deixava o 201 antigo intacto, os dois a 100%, somando 200% e
		// servindo um ou outro ao acaso. A validacao de percentual nao pegava,
		// porque confere o payload que chega, nao o que ja esta no banco.
		// ON DELETE CASCADE leva junto o response_model.
		if len(ep.Statuses) > 0 {
			keep := make([]interface{}, 0, len(ep.Statuses)+1)
			placeholders := ""
			keep = append(keep, urlConfigID)
			for i, st := range ep.Statuses {
				if i > 0 {
					placeholders += ", "
				}
				placeholders += fmt.Sprintf("$%d", i+2)
				keep = append(keep, st.HTTPStatus)
			}
			if _, err = tx.Exec(
				fmt.Sprintf(`DELETE FROM url_http_status WHERE url_id = $1 AND http_status NOT IN (%s)`, placeholders),
				keep...,
			); err != nil {
				response.SendResponse(w, http.StatusInternalServerError, "Failed to prune url_http_status", err.Error(), nil, false)
				return
			}
		}

		var statusResults []ScenarioStatusResult

		for _, st := range ep.Statuses {
			// Upsert url_http_status by (url_id, http_status)
			var httpStatusID int64
			err = tx.QueryRow(
				`SELECT id FROM url_http_status WHERE url_id = $1 AND http_status = $2 LIMIT 1`,
				urlConfigID, st.HTTPStatus,
			).Scan(&httpStatusID)
			if err == sql.ErrNoRows {
				err = tx.QueryRow(
					`INSERT INTO url_http_status (url_id, http_status, percentage) VALUES ($1, $2, $3) RETURNING id`,
					urlConfigID, st.HTTPStatus, st.Percentage,
				).Scan(&httpStatusID)
				if err != nil {
					response.SendResponse(w, http.StatusInternalServerError, "Failed to create url_http_status", err.Error(), nil, false)
					return
				}
			} else if err != nil {
				response.SendResponse(w, http.StatusInternalServerError, "Failed to query url_http_status", err.Error(), nil, false)
				return
			} else {
				if _, err = tx.Exec(
					`UPDATE url_http_status SET percentage = $1 WHERE id = $2`,
					st.Percentage, httpStatusID,
				); err != nil {
					response.SendResponse(w, http.StatusInternalServerError, "Failed to update url_http_status", err.Error(), nil, false)
					return
				}
			}

			// Upsert response_model by url_http_status_id (one model per status)
			var responseModelID int64
			err = tx.QueryRow(
				`SELECT id FROM response_model WHERE url_http_status_id = $1 LIMIT 1`,
				httpStatusID,
			).Scan(&responseModelID)
			if err == sql.ErrNoRows {
				err = tx.QueryRow(
					`INSERT INTO response_model (url_http_status_id, model) VALUES ($1, $2) RETURNING id`,
					httpStatusID, string(st.Model),
				).Scan(&responseModelID)
				if err != nil {
					response.SendResponse(w, http.StatusInternalServerError, "Failed to create response_model", err.Error(), nil, false)
					return
				}
			} else if err != nil {
				response.SendResponse(w, http.StatusInternalServerError, "Failed to query response_model", err.Error(), nil, false)
				return
			} else {
				if _, err = tx.Exec(
					`UPDATE response_model SET model = $1 WHERE id = $2`,
					string(st.Model), responseModelID,
				); err != nil {
					response.SendResponse(w, http.StatusInternalServerError, "Failed to update response_model", err.Error(), nil, false)
					return
				}
			}

			statusResults = append(statusResults, ScenarioStatusResult{
				URLHTTPStatusID: httpStatusID,
				ResponseModelID: responseModelID,
			})
		}

		endpointResults = append(endpointResults, ScenarioEndpointResult{
			URLConfigID: urlConfigID,
			Statuses:    statusResults,
		})
	}

	if err = tx.Commit(); err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to commit transaction", err.Error(), nil, false)
		return
	}

	result := ScenarioResult{
		ProjectID: projectID,
		Endpoints: endpointResults,
	}
	response.SendResponse(w, http.StatusOK, "Scenario created successfully", "", result, false)
}

// GetScenarioHandler handles GET /api/scenario/{project_id}.
// Reconstructs and returns the full scenario in the same format as POST.
func GetScenarioHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectIDStr := vars["project_id"]
	projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		response.SendResponse(w, http.StatusBadRequest, "Invalid project ID", err.Error(), nil, false)
		return
	}

	project, err := crud.Read("project", projectID)
	if err != nil {
		response.SendResponse(w, http.StatusNotFound, "Project not found", err.Error(), nil, false)
		return
	}

	urlConfigs, err := crud.List("url_config", map[string]interface{}{"project_id": projectID})
	if err != nil {
		response.SendResponse(w, http.StatusInternalServerError, "Failed to fetch url_configs", err.Error(), nil, false)
		return
	}

	type StatusOutput struct {
		HTTPStatus int             `json:"http_status"`
		Percentage int             `json:"percentage"`
		Model      json.RawMessage `json:"model"`
	}
	type EndpointOutput struct {
		Path        string         `json:"path"`
		Method      string         `json:"method"`
		Description string         `json:"description"`
		Statuses    []StatusOutput `json:"statuses"`
	}
	type ScenarioOutput struct {
		Project   map[string]interface{} `json:"project"`
		Endpoints []EndpointOutput       `json:"endpoints"`
	}

	var endpoints []EndpointOutput
	for _, uc := range urlConfigs {
		ucID, _ := uc["id"].(int64)

		httpStatuses, err := crud.List("url_http_status", map[string]interface{}{"url_id": ucID})
		if err != nil {
			response.SendResponse(w, http.StatusInternalServerError, "Failed to fetch http statuses", err.Error(), nil, false)
			return
		}

		var statusOutputs []StatusOutput
		for _, st := range httpStatuses {
			stID, _ := st["id"].(int64)

			models, err := crud.List("response_model", map[string]interface{}{"url_http_status_id": stID})
			if err != nil {
				response.SendResponse(w, http.StatusInternalServerError, "Failed to fetch response models", err.Error(), nil, false)
				return
			}

			var modelRaw json.RawMessage
			if len(models) > 0 {
				switch v := models[0]["model"].(type) {
				case []byte:
					modelRaw = json.RawMessage(v)
				case string:
					modelRaw = json.RawMessage(v)
				}
			}

			statusOutputs = append(statusOutputs, StatusOutput{
				HTTPStatus: int(st["http_status"].(int64)),
				Percentage: int(st["percentage"].(int64)),
				Model:      modelRaw,
			})
		}

		description := ""
		if d := uc["description"]; d != nil {
			switch v := d.(type) {
			case string:
				description = v
			case []byte:
				description = string(v)
			}
		}

		endpoints = append(endpoints, EndpointOutput{
			Path:        uc["path"].(string),
			Method:      uc["method"].(string),
			Description: description,
			Statuses:    statusOutputs,
		})
	}

	output := ScenarioOutput{
		Project:   project,
		Endpoints: endpoints,
	}
	response.SendResponse(w, http.StatusOK, "Scenario retrieved successfully", "", output, false)
}
