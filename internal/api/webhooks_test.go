package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFormWebhooks_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/form/555/webhooks")
		_ = json.NewEncoder(w).Encode(apiResponse[map[string]string]{
			ResponseCode: 200,
			Content: map[string]string{
				"0": "https://example.com/hook1",
				"1": "https://example.com/hook2",
			},
		})
	})

	webhooks, err := client.GetFormWebhooks("555")
	require.NoError(t, err)
	assert.Len(t, webhooks, 2)

	urls := map[string]bool{}
	for _, wh := range webhooks {
		urls[wh.URL] = true
	}
	assert.True(t, urls["https://example.com/hook1"])
	assert.True(t, urls["https://example.com/hook2"])
}

func TestGetFormWebhooks_Empty(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(apiResponse[map[string]string]{
			ResponseCode: 200,
			Content:      map[string]string{},
		})
	})

	webhooks, err := client.GetFormWebhooks("555")
	require.NoError(t, err)
	assert.Empty(t, webhooks)
}

func TestCreateFormWebhook_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/form/555/webhooks")
		w.WriteHeader(http.StatusOK)
	})

	err := client.CreateFormWebhook("555", "https://example.com/hook")
	require.NoError(t, err)
}

func TestCreateFormWebhook_Error(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	})

	err := client.CreateFormWebhook("555", "invalid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "400")
}

func TestDeleteFormWebhook_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Contains(t, r.URL.Path, "/form/555/webhooks/0")
		w.WriteHeader(http.StatusOK)
	})

	err := client.DeleteFormWebhook("555", "0")
	require.NoError(t, err)
}

func TestDeleteFormWebhook_Error(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	})

	err := client.DeleteFormWebhook("555", "99")
	require.Error(t, err)
}

func TestTestWebhook_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	code, err := TestWebhook(ts.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, code)
}

func TestMarshalTestPayload(t *testing.T) {
	data, err := MarshalTestPayload()
	require.NoError(t, err)

	var payload WebhookPayload
	err = json.Unmarshal(data, &payload)
	require.NoError(t, err)
	assert.True(t, payload.Test)
	assert.Equal(t, "jotform-cli", payload.Source)
	assert.NotEmpty(t, payload.Message)
}
