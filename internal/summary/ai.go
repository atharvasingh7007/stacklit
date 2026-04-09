package summary

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/glincker/stacklit/internal/schema"
)

const (
	claudeAPIURL   = "https://api.anthropic.com/v1/messages"
	claudeModel    = "claude-sonnet-4-20250514"
	claudeVersion  = "2023-06-01"
	maxTokens      = 500
	systemPrompt   = "You are a senior software architect. Summarize this codebase architecture in 2-3 concise paragraphs. Focus on the overall structure, key patterns, and how data flows. Be specific to this codebase, not generic."
)

// indexSnapshot is the subset of the index sent to the API.
type indexSnapshot struct {
	Project      schema.Project               `json:"project"`
	Tech         schema.Tech                  `json:"tech"`
	Modules      map[string]schema.ModuleInfo `json:"modules"`
	Dependencies schema.Dependencies          `json:"dependencies"`
	Entrypoints  []string                     `json:"entrypoints"`
}

// claudeRequest is the request body for the Anthropic Messages API.
type claudeRequest struct {
	Model     string              `json:"model"`
	MaxTokens int                 `json:"max_tokens"`
	System    string              `json:"system"`
	Messages  []map[string]string `json:"messages"`
}

// claudeResponse is the top-level response from the Anthropic Messages API.
type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Generate calls the Anthropic Claude API and returns a 2-3 paragraph narrative
// summary of the codebase architecture described by idx.
func Generate(idx *schema.Index) (string, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY not set. Set it to generate AI summaries")
	}

	snapshot := indexSnapshot{
		Project:      idx.Project,
		Tech:         idx.Tech,
		Modules:      idx.Modules,
		Dependencies: idx.Dependencies,
		Entrypoints:  idx.Structure.Entrypoints,
	}

	userMsg, err := json.Marshal(snapshot)
	if err != nil {
		return "", fmt.Errorf("marshalling index snapshot: %w", err)
	}

	return callClaude(apiKey, systemPrompt, string(userMsg))
}

func callClaude(apiKey, system, userMessage string) (string, error) {
	body := claudeRequest{
		Model:     claudeModel,
		MaxTokens: maxTokens,
		System:    system,
		Messages: []map[string]string{
			{"role": "user", "content": userMessage},
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshalling request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, claudeAPIURL, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", claudeVersion)
	req.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling Anthropic API: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	var result claudeResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("API error (%s): %s", result.Error.Type, result.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(data))
	}

	for _, block := range result.Content {
		if block.Type == "text" {
			return block.Text, nil
		}
	}

	return "", fmt.Errorf("no text content in API response")
}
