package translate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// libreTranslator calls a (self-hosted or public) LibreTranslate instance.
type libreTranslator struct {
	endpoint string
	apiKey   string // optional; some instances require it
	client   *http.Client
}

func newLibreTranslator(endpoint, apiKey string) *libreTranslator {
	return &libreTranslator{
		endpoint: strings.TrimRight(endpoint, "/"),
		apiKey:   apiKey,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

type libreRequest struct {
	Q      string `json:"q"`
	Source string `json:"source"`
	Target string `json:"target"`
	Format string `json:"format"`
	APIKey string `json:"api_key,omitempty"`
}

type libreResponse struct {
	TranslatedText   string `json:"translatedText"`
	DetectedLanguage *struct {
		Language string `json:"language"`
	} `json:"detectedLanguage,omitempty"`
	Error string `json:"error,omitempty"`
}

func (t *libreTranslator) Translate(ctx context.Context, text, targetLang string) (Result, error) {
	payload := libreRequest{
		Q:      text,
		Source: "auto",
		Target: strings.ToLower(targetLang),
		Format: "text",
		APIKey: t.apiKey,
	}
	buf, err := json.Marshal(payload)
	if err != nil {
		return Result{}, fmt.Errorf("translate/libre: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.endpoint+"/translate", bytes.NewReader(buf))
	if err != nil {
		return Result{}, fmt.Errorf("translate/libre: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("translate/libre: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("translate/libre: unexpected status %d", resp.StatusCode)
	}

	var body libreResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return Result{}, fmt.Errorf("translate/libre: decode response: %w", err)
	}
	if body.Error != "" {
		return Result{}, fmt.Errorf("translate/libre: %s", body.Error)
	}
	if body.TranslatedText == "" {
		return Result{}, fmt.Errorf("translate/libre: empty translation response")
	}

	res := Result{TranslatedText: body.TranslatedText}
	if body.DetectedLanguage != nil {
		res.DetectedSourceLang = normalizeLang(body.DetectedLanguage.Language)
	}
	return res, nil
}
