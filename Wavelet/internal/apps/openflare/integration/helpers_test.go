// Copyright 2026 Arctel.net
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Rain-kl/Wavelet/internal/apps/openflare/compat"
	"github.com/stretchr/testify/require"
)

func decodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder) compat.Envelope {
	t.Helper()

	var envelope compat.Envelope
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	return envelope
}

func unmarshalEnvelopeData(t *testing.T, data any, target any) {
	t.Helper()

	payload, err := json.Marshal(data)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(payload, target))
}

func unmarshalEnvelopeMap(t *testing.T, data any) map[string]any {
	t.Helper()

	var result map[string]any
	unmarshalEnvelopeData(t, data, &result)
	return result
}

func unmarshalEnvelopeSlice(t *testing.T, data any) []any {
	t.Helper()

	var result []any
	unmarshalEnvelopeData(t, data, &result)
	return result
}

func performJSONRequest(
	t *testing.T,
	engine http.Handler,
	method, path string,
	body any,
	headers map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		require.NoError(t, err)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func performLegacyRequest(
	t *testing.T,
	engine http.Handler,
	method, path string,
	body any,
	headers map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()
	return performJSONRequest(t, engine, method, path, body, headers)
}

func adminAuthHeaders(token string) map[string]string {
	return map[string]string{
		"X-Access-Token": token,
	}
}
