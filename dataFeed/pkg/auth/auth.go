// Package auth provides authentication functionality
package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"datafeed/pkg/config"
)

// Login authenticates to the remote service and returns a token
func Login(cfg *config.Config) (string, error) {
	payload := map[string]string{
		"loginId":  cfg.Username,
		"password": cfg.Password,
		"deviceId": "d72dc7b5-14d2-4896-83e4-cfc7a3fd625f", // Replace with actual device ID if needed
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(cfg.LoginURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("login failed: " + resp.Status)
	}

	// Extract token from JSON response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// Try to extract token from nested body (e.g., result["data"]["token"])
	if data, ok := result["data"].(map[string]interface{}); ok {
		if token, ok := data["accessToken"].(string); ok {
			return token, nil
		}
		if msg, ok := data["errorMessage"].(string); ok && msg != "" {
			return "", errors.New(msg)
		}
	}
	return "", errors.New("token not found in login response")
}
