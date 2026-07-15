package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// googleTranslator calls the Google Cloud Translation API (v2), which
// authenticates with a simple API key query parameter.
type googleTranslator struct {
	apiKey string
	client *http.Client
}

func newGoogleTranslator(apiKey string) *googleTranslator {
	return &googleTranslator{
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type googleResponse struct {
	Data struct {
		Translations []struct {
			TranslatedText         string `json:"translatedText"`
			DetectedSourceLanguage string `json:"detectedSourceLanguage"`
		} `json:"translations"`
	} `json:"data"`
}

func (t *googleTranslator) Translate(ctx context.Context, text, targetLang string) (Result, error) {
	form := url.Values{}
	form.Set("q", text)
	form.Set("target", strings.ToLower(targetLang))
	form.Set("format", "text")

	endpoint := "https://translation.googleapis.com/language/translate/v2?key=" + url.QueryEscape(t.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return Result{}, fmt.Errorf("translate/google: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.client.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("translate/google: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("translate/google: unexpected status %d", resp.StatusCode)
	}

	var body googleResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return Result{}, fmt.Errorf("translate/google: decode response: %w", err)
	}
	if len(body.Data.Translations) == 0 {
		return Result{}, fmt.Errorf("translate/google: empty translation response")
	}

	return Result{
		TranslatedText:     body.Data.Translations[0].TranslatedText,
		DetectedSourceLang: normalizeLang(body.Data.Translations[0].DetectedSourceLanguage),
	}, nil
}
