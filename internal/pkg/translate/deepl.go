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

// deeplTranslator calls the DeepL API (free tier endpoint).
type deeplTranslator struct {
	apiKey string
	client *http.Client
}

func newDeepLTranslator(apiKey string) *deeplTranslator {
	return &deeplTranslator{
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// deeplEndpoint returns the correct host based on the API key suffix. DeepL
// free keys end in ":fx" and use api-free; paid keys use api.
func (t *deeplTranslator) deeplEndpoint() string {
	if strings.HasSuffix(t.apiKey, ":fx") {
		return "https://api-free.deepl.com/v2/translate"
	}
	return "https://api.deepl.com/v2/translate"
}

type deeplResponse struct {
	Translations []struct {
		DetectedSourceLanguage string `json:"detected_source_language"`
		Text                   string `json:"text"`
	} `json:"translations"`
}

func (t *deeplTranslator) Translate(ctx context.Context, text, targetLang string) (Result, error) {
	form := url.Values{}
	form.Set("text", text)
	form.Set("target_lang", strings.ToUpper(targetLang))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.deeplEndpoint(), strings.NewReader(form.Encode()))
	if err != nil {
		return Result{}, fmt.Errorf("translate/deepl: build request: %w", err)
	}
	req.Header.Set("Authorization", "DeepL-Auth-Key "+t.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.client.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("translate/deepl: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{}, fmt.Errorf("translate/deepl: unexpected status %d", resp.StatusCode)
	}

	var body deeplResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return Result{}, fmt.Errorf("translate/deepl: decode response: %w", err)
	}
	if len(body.Translations) == 0 {
		return Result{}, fmt.Errorf("translate/deepl: empty translation response")
	}

	return Result{
		TranslatedText:     body.Translations[0].Text,
		DetectedSourceLang: normalizeLang(body.Translations[0].DetectedSourceLanguage),
	}, nil
}
