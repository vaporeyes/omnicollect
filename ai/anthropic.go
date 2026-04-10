// ABOUTME: Anthropic Messages API client for vision-based image analysis.
// ABOUTME: Sends base64 images to the Anthropic API and extracts text responses.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// AnthropicProvider calls the Anthropic Messages API directly.
type AnthropicProvider struct {
	apiKey string
	model  string
}

// anthropicRequest is the Messages API request body.
type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string        `json:"role"`
	Content []interface{} `json:"content"`
}

type anthropicImageBlock struct {
	Type   string               `json:"type"`
	Source anthropicImageSource `json:"source"`
}

type anthropicImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type anthropicTextBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// anthropicResponse is the Messages API response body.
type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// AnalyzeImage sends an image and prompt to the Anthropic Messages API.
func (p *AnthropicProvider) AnalyzeImage(ctx context.Context, imageBase64 string, prompt string) (string, error) {
	reqBody := anthropicRequest{
		Model:     p.model,
		MaxTokens: 4096,
		Messages: []anthropicMessage{
			{
				Role: "user",
				Content: []interface{}{
					anthropicImageBlock{
						Type: "image",
						Source: anthropicImageSource{
							Type:      "base64",
							MediaType: "image/jpeg",
							Data:      imageBase64,
						},
					},
					anthropicTextBlock{
						Type: "text",
						Text: prompt,
					},
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling Anthropic API: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Anthropic API returned %d: %s", resp.StatusCode, string(respBytes))
	}

	var result anthropicResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("Anthropic API error: %s", result.Error.Message)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("Anthropic API returned empty content")
	}

	return result.Content[0].Text, nil
}
