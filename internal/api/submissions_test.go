package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSubmissions_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/form/111/submissions")
		assert.Contains(t, r.URL.RawQuery, "offset=0")
		assert.Contains(t, r.URL.RawQuery, "limit=20")
		assert.Contains(t, r.URL.RawQuery, "orderby=created_at")
		assert.Contains(t, r.URL.RawQuery, "direction=DESC")

		_ = json.NewEncoder(w).Encode(apiResponse[[]Submission]{
			ResponseCode: 200,
			Content: []Submission{
				{ID: "1001", FormID: "111", Status: "ACTIVE", CreatedAt: "2024-01-01"},
				{ID: "1002", FormID: "111", Status: "ACTIVE", CreatedAt: "2024-01-02"},
			},
		})
	})

	subs, err := client.GetSubmissions("111", 0, 20, "created_at", "DESC")
	require.NoError(t, err)
	assert.Len(t, subs, 2)
	assert.Equal(t, "1001", subs[0].ID)
	assert.Equal(t, "111", subs[0].FormID)
}

func TestGetSubmissions_Empty(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(apiResponse[[]Submission]{
			ResponseCode: 200,
			Content:      []Submission{},
		})
	})

	subs, err := client.GetSubmissions("111", 0, 20, "created_at", "DESC")
	require.NoError(t, err)
	assert.Empty(t, subs)
}

func TestGetSubmissions_Error(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server error"))
	})

	_, err := client.GetSubmissions("111", 0, 20, "created_at", "DESC")
	require.Error(t, err)
}
