package requests

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const githubGraphQLAPI = "https://api.github.com/graphql"

func Request(token, query string) (*http.Response, error) {
	payload := struct {
		Query string `json:"query"`
	}{
		Query: query,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", githubGraphQLAPI, bytes.NewBuffer(payloadJSON)) // Use bytes.NewBuffer
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json") // Important: Set Content-Type header

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	return resp, err
}
