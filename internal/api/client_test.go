package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	c := New("test-key")
	c.BaseURL = ts.URL
	return ts, c
}

func TestGetUser_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.RawQuery, "apiKey=test-key")
		_ = json.NewEncoder(w).Encode(apiResponse[User]{
			ResponseCode: 200,
			Content:      User{Username: "testuser", Email: "test@example.com", AccountType: "FREE"},
		})
	})

	user, err := client.GetUser()
	require.NoError(t, err)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "FREE", user.AccountType)
}

func TestGetUser_Unauthorized(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	_, err := client.GetUser()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid API key")
}

func TestListForms_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/user/forms")
		_ = json.NewEncoder(w).Encode(apiResponse[[]Form]{
			ResponseCode: 200,
			Content: []Form{
				{ID: "111", Title: "Survey A"},
				{ID: "222", Title: "Survey B"},
			},
		})
	})

	forms, err := client.ListForms(0, 10)
	require.NoError(t, err)
	assert.Len(t, forms, 2)
	assert.Equal(t, "Survey A", forms[0].Title)
}

func TestListForms_Empty(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(apiResponse[[]Form]{
			ResponseCode: 200,
			Content:      []Form{},
		})
	})

	forms, err := client.ListForms(0, 10)
	require.NoError(t, err)
	assert.Empty(t, forms)
}

func TestGetForm_Success(t *testing.T) {
	seenBase := false
	seenQuestions := false
	seenProperties := false

	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/form/123":
			seenBase = true
			_ = json.NewEncoder(w).Encode(apiResponse[FormProperties]{
				ResponseCode: 200,
				Content:      FormProperties{ID: "123", Title: "My Form"},
			})
		case "/form/123/questions":
			seenQuestions = true
			_ = json.NewEncoder(w).Encode(apiResponse[map[string]interface{}]{
				ResponseCode: 200,
				Content: map[string]interface{}{
					"1": map[string]interface{}{"type": "control_textbox", "text": "Name"},
				},
			})
		case "/form/123/properties":
			seenProperties = true
			_ = json.NewEncoder(w).Encode(apiResponse[map[string]interface{}]{
				ResponseCode: 200,
				Content:      map[string]interface{}{"title": "My Form"},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	form, err := client.GetForm("123")
	require.NoError(t, err)
	assert.Equal(t, "My Form", form.Title)
	assert.NotEmpty(t, form.Questions)
	assert.True(t, seenBase)
	assert.True(t, seenQuestions)
	assert.True(t, seenProperties)
}
