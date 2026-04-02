package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- Count.UnmarshalJSON ----

func TestCount_UnmarshalJSON_String(t *testing.T) {
	var f Form
	require.NoError(t, json.Unmarshal([]byte(`{"count":"42"}`), &f))
	assert.Equal(t, Count("42"), f.Count)
}

func TestCount_UnmarshalJSON_Number(t *testing.T) {
	var f Form
	require.NoError(t, json.Unmarshal([]byte(`{"count":17}`), &f))
	assert.Equal(t, Count("17"), f.Count)
}

func TestCount_UnmarshalJSON_Null(t *testing.T) {
	var f Form
	require.NoError(t, json.Unmarshal([]byte(`{"count":null}`), &f))
	assert.Equal(t, Count(""), f.Count)
}

func TestCount_UnmarshalJSON_Empty(t *testing.T) {
	var f Form
	require.NoError(t, json.Unmarshal([]byte(`{}`), &f))
	assert.Equal(t, Count(""), f.Count)
}

// ---- flattenFormSchema ----

func TestFlattenFormSchema_Title(t *testing.T) {
	schema := map[string]interface{}{"title": "My Form"}
	values := flattenFormSchema(schema)
	assert.Equal(t, "My Form", values.Get("properties[title]"))
}

func TestFlattenFormSchema_Properties(t *testing.T) {
	schema := map[string]interface{}{
		"properties": map[string]interface{}{
			"font":      "Arial",
			"highlight": "blue",
		},
	}
	values := flattenFormSchema(schema)
	assert.Equal(t, "Arial", values.Get("properties[font]"))
	assert.Equal(t, "blue", values.Get("properties[highlight]"))
}

func TestFlattenFormSchema_Questions(t *testing.T) {
	schema := map[string]interface{}{
		"questions": map[string]interface{}{
			"1": map[string]interface{}{
				"type": "control_textbox",
				"text": "Name",
			},
		},
	}
	values := flattenFormSchema(schema)
	assert.Equal(t, "control_textbox", values.Get("questions[1][type]"))
	assert.Equal(t, "Name", values.Get("questions[1][text]"))
}

func TestFlattenFormSchema_Empty(t *testing.T) {
	values := flattenFormSchema(map[string]interface{}{})
	assert.Empty(t, values)
}

func TestFlattenFormSchema_TitleTakesPrecedence(t *testing.T) {
	// title key at top level sets properties[title]
	schema := map[string]interface{}{
		"title": "Top Level Title",
		"properties": map[string]interface{}{
			"font": "Verdana",
		},
	}
	values := flattenFormSchema(schema)
	assert.Equal(t, "Top Level Title", values.Get("properties[title]"))
	assert.Equal(t, "Verdana", values.Get("properties[font]"))
}

// ---- CreateForm ----

func TestCreateForm_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/user/forms":
			_ = json.NewEncoder(w).Encode(apiResponse[Form]{
				ResponseCode: 200,
				Content:      Form{ID: "new-form-id", Title: "New Form"},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/form/new-form-id/properties":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPut && r.URL.Path == "/form/new-form-id/questions":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(apiResponse[FormProperties]{
				Content: FormProperties{ID: "new-form-id", Title: "New Form"},
			})
		}
	}))
	defer ts.Close()

	c := New("test-key")
	c.BaseURL = ts.URL

	schema := map[string]interface{}{
		"title": "New Form",
		"questions": map[string]interface{}{
			"1": map[string]interface{}{"type": "control_textbox", "text": "Name"},
		},
	}

	form, err := c.CreateForm(schema)
	require.NoError(t, err)
	assert.Equal(t, "new-form-id", form.ID)
}

func TestCreateForm_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server error"))
	}))
	defer ts.Close()

	c := New("test-key")
	c.BaseURL = ts.URL

	_, err := c.CreateForm(map[string]interface{}{"title": "Form"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestCreateForm_Unauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer ts.Close()

	c := New("bad-key")
	c.BaseURL = ts.URL

	_, err := c.CreateForm(map[string]interface{}{"title": "Form"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

// ---- UpdateForm ----

func TestUpdateForm_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/form/123/properties":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPut && r.URL.Path == "/form/123/questions":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/form/123":
			_ = json.NewEncoder(w).Encode(apiResponse[FormProperties]{
				Content: FormProperties{ID: "123", Title: "Updated Form"},
			})
		case r.URL.Path == "/form/123/questions":
			_ = json.NewEncoder(w).Encode(apiResponse[map[string]interface{}]{
				Content: map[string]interface{}{},
			})
		case r.URL.Path == "/form/123/properties":
			_ = json.NewEncoder(w).Encode(apiResponse[map[string]interface{}]{
				Content: map[string]interface{}{"title": "Updated Form"},
			})
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer ts.Close()

	c := New("test-key")
	c.BaseURL = ts.URL

	schema := map[string]interface{}{
		"title": "Updated Form",
		"properties": map[string]interface{}{
			"font": "Arial",
		},
	}

	form, err := c.UpdateForm("123", schema)
	require.NoError(t, err)
	assert.Equal(t, "123", form.ID)
}

func TestUpdateForm_PropertiesError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/form/123/properties" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("properties error"))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := New("test-key")
	c.BaseURL = ts.URL

	_, err := c.UpdateForm("123", map[string]interface{}{
		"properties": map[string]interface{}{"font": "Arial"},
	})
	require.Error(t, err)
}

// ---- DeleteForm ----

func TestDeleteForm_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Contains(t, r.URL.Path, "/form/456")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := New("test-key")
	c.BaseURL = ts.URL

	require.NoError(t, c.DeleteForm("456"))
}

func TestDeleteForm_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := New("test-key")
	c.BaseURL = ts.URL

	err := c.DeleteForm("999")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
}

func TestDeleteForm_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := New("test-key")
	c.BaseURL = ts.URL

	err := c.DeleteForm("1")
	require.Error(t, err)
}

// ---- setFormProperties (via UpdateForm with no questions) ----

func TestSetFormProperties_NoPropertiesInSchema(t *testing.T) {
	// Schema with no properties or title — setFormProperties should skip the POST
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/form/1/properties" {
			called = true
		}
		// GetForm calls for UpdateForm
		_ = json.NewEncoder(w).Encode(apiResponse[FormProperties]{
			Content: FormProperties{ID: "1"},
		})
	}))
	defer ts.Close()

	c := New("test-key")
	c.BaseURL = ts.URL

	// Schema with no title and no properties — setFormProperties should be a no-op
	_, _ = c.UpdateForm("1", map[string]interface{}{})
	assert.False(t, called, "POST /properties should not be called when schema has no properties")
}
