// ABOUTME: Shared deadlines, response budgets, and image MIME validation for AI requests.
// ABOUTME: Keeps external provider failures bounded and rejects unsupported image payloads.
package ai

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

var providerHTTPClient = &http.Client{Timeout: 60 * time.Second}

func validateMediaType(value string) error {
	switch value {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return nil
	}
	return fmt.Errorf("unsupported image MIME type")
}
func readProviderResponse(reader io.Reader) ([]byte, error) {
	const limit = 2 << 20
	data, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if len(data) > limit {
		return nil, fmt.Errorf("AI response exceeds size limit")
	}
	return data, nil
}
