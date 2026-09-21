package handler

import (
	"testing"
)

func TestValidateScenarioInput_PercentageExceeds100(t *testing.T) {
	input := ScenarioInput{
		Project: ScenarioProjectInput{Name: "my-project"},
		Endpoints: []ScenarioEndpointInput{
			{
				Path:   "/test",
				Method: "GET",
				Statuses: []ScenarioStatusInput{
					{HTTPStatus: 200, Percentage: 60},
					{HTTPStatus: 500, Percentage: 50}, // sum = 110
				},
			},
		},
	}
	if err := validateScenarioInput(input); err == nil {
		t.Error("expected error when sum of percentages > 100")
	}
}

func TestValidateScenarioInput_MissingProjectName(t *testing.T) {
	input := ScenarioInput{
		Project: ScenarioProjectInput{Name: ""},
		Endpoints: []ScenarioEndpointInput{
			{
				Path:   "/test",
				Method: "GET",
				Statuses: []ScenarioStatusInput{
					{HTTPStatus: 200, Percentage: 100},
				},
			},
		},
	}
	if err := validateScenarioInput(input); err == nil {
		t.Error("expected error for missing project name")
	}
}

func TestValidateScenarioInput_EmptyStatuses(t *testing.T) {
	input := ScenarioInput{
		Project: ScenarioProjectInput{Name: "proj"},
		Endpoints: []ScenarioEndpointInput{
			{Path: "/test", Method: "GET", Statuses: nil},
		},
	}
	if err := validateScenarioInput(input); err == nil {
		t.Error("expected error when statuses is empty")
	}
}

func TestValidateScenarioInput_Valid(t *testing.T) {
	input := ScenarioInput{
		Project: ScenarioProjectInput{Name: "proj"},
		Endpoints: []ScenarioEndpointInput{
			{
				Path:   "/test",
				Method: "GET",
				Statuses: []ScenarioStatusInput{
					{HTTPStatus: 200, Percentage: 80},
					{HTTPStatus: 500, Percentage: 20},
				},
			},
		},
	}
	if err := validateScenarioInput(input); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}
