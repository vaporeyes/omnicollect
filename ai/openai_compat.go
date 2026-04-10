// ABOUTME: OpenAI-compatible API client for vision-based image analysis.
// ABOUTME: Works with OpenRouter, Google Gemini, and any OpenAI-format endpoint.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OpenAICompatProvider calls any OpenAI-compatible chat completions endpoint.
type OpenAICompatProvider struct {
	apiKey  string
	model   string
	baseURL string
}

// openaiRequest is the chat completions request body.
type openaiRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []openaiMessage `json:"messages"`
}

type openaiMessage struct {
	Role    string        `json:"role"`
	Content []interface{} `json:"content"`
}

type openaiImageURLBlock struct {
	Type     string         `json:"type"`
	ImageURL openaiImageURL `json:"image_url"`
}

type openaiImageURL struct {
	URL string `json:"url"`
}

type openaiTextBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// openaiResponse is the chat completions response body.
type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// AnalyzeImage sends an image and prompt to an OpenAI-compatible endpoint.
func (p *OpenAICompatProvider) AnalyzeImage(ctx context.Context, imageBase64 string, prompt string) (string, error) {
	reqBody := openaiRequest{
		Model:     p.model,
		MaxTokens: 4096,
		Messages: []openaiMessage{
			{
				Role: "user",
				Content: []interface{}{
					openaiImageURLBlock{
						Type: "image_url",
						ImageURL: openaiImageURL{
							URL: "data:image/jpeg;base64," + imageBase64,
						},
					},
					openaiTextBlock{
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

	url := strings.TrimRight(p.baseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling AI API: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI API returned %d: %s", resp.StatusCode, string(respBytes))
	}

	var result openaiResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("AI API error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("AI API returned no choices")
	}

	return result.Choices[0].Message.Content, nil
}
