package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/clock/service"
	"github.com/stretchr/testify/assert"
)

func TestWriteServiceError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"validation error is a 400", &service.ValidationError{Msg: "unsupported timezone"}, http.StatusBadRequest},
		{"wrapped validation error is a 400", fmt.Errorf("create: %w", &service.ValidationError{Msg: "x"}), http.StatusBadRequest},
		{"not found is a 404", service.ErrNotFound, http.StatusNotFound},
		{"wrapped not found is a 404", fmt.Errorf("get: %w", service.ErrNotFound), http.StatusNotFound},
		{"unknown error is a 500", errors.New("db down"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			writeServiceError(rec, tt.err, "boom")
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestWriteServiceError_DoesNotLeakInternals(t *testing.T) {
	rec := httptest.NewRecorder()
	writeServiceError(rec, errors.New("pq: password authentication failed"), "Failed to create clock channel")

	assert.NotContains(t, rec.Body.String(), "password")
	assert.Contains(t, rec.Body.String(), "Failed to create clock channel")
}

func TestParseOptionalID(t *testing.T) {
	str := func(s string) *string { return &s }
	id := int64(123456789012345678)

	tests := []struct {
		name    string
		in      *string
		want    *int64
		wantErr bool
	}{
		{"nil is absent", nil, nil, false},
		{"empty is absent", str(""), nil, false},
		{"valid snowflake", str("123456789012345678"), &id, false},
		{"not a number", str("abc"), nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOptionalID(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
