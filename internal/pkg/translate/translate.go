// Package translate provides a small abstraction over external translation
// engines (DeepL, Google Cloud Translation, LibreTranslate). Each engine
// implements the Translator interface, and NewTranslator builds the correct
// implementation from a guild's stored configuration.
package translate

import (
	"context"
	"fmt"
	"strings"
)

// Supported provider identifiers (stored in translation_config.provider).
const (
	ProviderDeepL          = "deepl"
	ProviderGoogle         = "google"
	ProviderLibreTranslate = "libretranslate"
)

// Result is the outcome of a translation request.
type Result struct {
	// TranslatedText is the text rendered in the target language.
	TranslatedText string
	// DetectedSourceLang is the source language the engine detected, as a
	// lowercase ISO code (e.g. "en"). May be empty if the engine does not
	// report it.
	DetectedSourceLang string
}

// Translator translates text into a target language.
type Translator interface {
	// Translate renders text into targetLang (an ISO 639-1 code such as "en",
	// "es", "ja"). Implementations auto-detect the source language.
	Translate(ctx context.Context, text, targetLang string) (Result, error)
}

// NewTranslator builds a Translator for the given provider. apiKey is the
// engine credential (may be empty for keyless self-hosted LibreTranslate);
// endpoint is the base URL for self-hosted engines and is ignored otherwise.
func NewTranslator(provider, apiKey, endpoint string) (Translator, error) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case ProviderDeepL, "":
		if apiKey == "" {
			return nil, fmt.Errorf("translate: DeepL requires an API key")
		}
		return newDeepLTranslator(apiKey), nil
	case ProviderGoogle:
		if apiKey == "" {
			return nil, fmt.Errorf("translate: Google Translation requires an API key")
		}
		return newGoogleTranslator(apiKey), nil
	case ProviderLibreTranslate:
		if endpoint == "" {
			return nil, fmt.Errorf("translate: LibreTranslate requires an endpoint URL")
		}
		return newLibreTranslator(endpoint, apiKey), nil
	default:
		return nil, fmt.Errorf("translate: unknown provider %q", provider)
	}
}

// normalizeLang lowercases a language code and strips any region suffix
// (e.g. "EN-GB" -> "en") for consistent comparison against a target language.
func normalizeLang(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	if i := strings.IndexAny(code, "-_"); i > 0 {
		code = code[:i]
	}
	return code
}
