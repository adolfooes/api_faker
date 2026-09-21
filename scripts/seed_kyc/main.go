package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const defaultAPIURL = "http://localhost:8080"

func loadEnvFile(path string) map[string]string {
	env := make(map[string]string)
	data, err := os.ReadFile(path)
	if err != nil {
		return env
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			env[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return env
}

func getEnv(fileEnv map[string]string, key, defaultVal string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	if v, ok := fileEnv[key]; ok && v != "" {
		return v
	}
	return defaultVal
}

func postJSON(url string, body interface{}, token string) (int, []byte, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return 0, nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, respBody, nil
}

func main() {
	env := loadEnvFile("config/.env")
	apiURL := getEnv(env, "API_URL", defaultAPIURL)
	email := getEnv(env, "SEED_EMAIL", "seed@apifaker.dev")
	password := getEnv(env, "SEED_PASSWORD", "Seed@1234!")

	// Create account if it doesn't exist yet.
	code, _, err := postJSON(apiURL+"/account", map[string]string{"email": email, "password": password}, "")
	if err != nil {
		log.Fatalf("create account request failed: %v", err)
	}
	if code != 201 && code != 409 {
		log.Fatalf("create account returned unexpected status %d", code)
	}

	// Login.
	code, body, err := postJSON(apiURL+"/login", map[string]string{"email": email, "password": password}, "")
	if err != nil {
		log.Fatalf("login request failed: %v", err)
	}
	if code != 200 {
		log.Fatalf("login failed (status %d): %s", code, body)
	}
	var loginResp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &loginResp); err != nil || loginResp.Data.Token == "" {
		log.Fatalf("parse login response failed: %v — body: %s", err, body)
	}
	token := loginResp.Data.Token
	fmt.Printf("Logged in as %s\n", email)

	// Read and parse scenario file.
	scenarioBytes, err := os.ReadFile("scripts/kyc-scenario.json")
	if err != nil {
		log.Fatalf("read scenario file: %v", err)
	}
	var scenario interface{}
	if err := json.Unmarshal(scenarioBytes, &scenario); err != nil {
		log.Fatalf("parse scenario file: %v", err)
	}

	// POST scenario (idempotent — upserts by project name + path + method).
	code, respBody, err := postJSON(apiURL+"/api/scenario", scenario, token)
	if err != nil {
		log.Fatalf("post scenario request failed: %v", err)
	}
	if code != 200 {
		log.Fatalf("post scenario returned status %d: %s", code, respBody)
	}
	fmt.Println("KYC scenario seeded successfully.")
}
