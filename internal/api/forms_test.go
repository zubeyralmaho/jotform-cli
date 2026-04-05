package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCount_UnmarshalJSON_String(t *testing.T) {
	var c Count
	err := c.UnmarshalJSON([]byte(`"42"`))
	require.NoError(t, err)
	assert.Equal(t, Count("42"), c)
}

func TestCount_UnmarshalJSON_Number(t *testing.T) {
	var c Count
	err := c.UnmarshalJSON([]byte(`42`))
	require.NoError(t, err)
	assert.Equal(t, Count("42"), c)
}

func TestCount_UnmarshalJSON_Null(t *testing.T) {
	var c Count
	err := c.UnmarshalJSON([]byte(`null`))
	require.NoError(t, err)
	assert.Equal(t, Count(""), c)
}

func TestCount_UnmarshalJSON_Empty(t *testing.T) {
	var c Count
	err := c.UnmarshalJSON([]byte{})
	require.NoError(t, err)
	assert.Equal(t, Count(""), c)
}

func TestCount_UnmarshalJSON_Invalid(t *testing.T) {
	var c Count
	err := c.UnmarshalJSON([]byte(`[1,2]`))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid count value")
}

func TestCreateForm_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/user/forms" && r.Method == http.MethodPost:
			_ = json.NewEncoder(w).Encode(apiResponse[Form]{
				ResponseCode: 200,
				Content:      Form{ID: "999", Title: "New Form"},
			})
		case r.URL.Path == "/form/999/properties" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"responseCode":200}`))
		case r.URL.Path == "/form/999/questions" && r.Method == http.MethodPut:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"responseCode":200}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	c := New("test-key")
	c.BaseURL = ts.URL

	schema := map[string]interface{}{
		"title": "New Form",
		"properties": map[string]interface{}{
			"title": "New Form",
		},
		"questions": map[string]interface{}{
			"1": map[string]interface{}{"type": "control_head", "text": "Header"},
		},
	}

	form, err := c.CreateForm(schema)
	require.NoError(t, err)
	assert.Equal(t, "999", form.ID)
}

func TestCreateForm_Unauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"unauthorized"}`))
	}))
	defer ts.Close()

	c := New("bad-key")
	c.BaseURL = ts.URL

	_, err := c.CreateForm(map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

func TestDeleteForm_Success(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Contains(t, r.URL.Path, "/form/456")
		w.WriteHeader(http.StatusOK)
	})

	err := client.DeleteForm("456")
	require.NoError(t, err)
}

func TestDeleteForm_Error(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	err := client.DeleteForm("456")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
}

func TestUpdateForm_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/form/100/properties" && r.Method == http.MethodPost:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"responseCode":200}`))
		case r.URL.Path == "/form/100/questions" && r.Method == http.MethodPut:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"responseCode":200}`))
		case r.URL.Path == "/form/100":
			_ = json.NewEncoder(w).Encode(apiResponse[FormProperties]{
				ResponseCode: 200,
				Content:      FormProperties{ID: "100", Title: "Updated"},
			})
		case r.URL.Path == "/form/100/questions":
			_ = json.NewEncoder(w).Encode(apiResponse[map[string]interface{}]{
				ResponseCode: 200,
				Content:      map[string]interface{}{},
			})
		case r.URL.Path == "/form/100/properties":
			_ = json.NewEncoder(w).Encode(apiResponse[map[string]interface{}]{
				ResponseCode: 200,
				Content:      map[string]interface{}{"title": "Updated"},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	c := New("test-key")
	c.BaseURL = ts.URL

	schema := map[string]interface{}{
		"properties": map[string]interface{}{"title": "Updated"},
	}

	form, err := c.UpdateForm("100", schema)
	require.NoError(t, err)
	assert.Equal(t, "100", form.ID)
	assert.Equal(t, "Updated", form.Title)
}

func TestFlattenFormSchema(t *testing.T) {
	schema := map[string]interface{}{
		"title": "My Form",
		"properties": map[string]interface{}{
			"activeRedirect": "thankYouPage",
		},
	}

	values := flattenFormSchema(schema)
	assert.Equal(t, "My Form", values.Get("properties[title]"))
	assert.Equal(t, "thankYouPage", values.Get("properties[activeRedirect]"))
}

func TestFlattenFormSchema_NoProperties(t *testing.T) {
	schema := map[string]interface{}{
		"title": "Only Title",
	}

	values := flattenFormSchema(schema)
	assert.Equal(t, "Only Title", values.Get("properties[title]"))
}

func TestFlattenFormSchema_EmptyTitle(t *testing.T) {
	schema := map[string]interface{}{
		"title": "",
	}

	values := flattenFormSchema(schema)
	assert.Equal(t, "", values.Get("properties[title]"))
}

func TestSetFormQuestions_NoQuestions(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not make any request when no questions")
	})

	err := client.setFormQuestions("123", map[string]interface{}{})
	require.NoError(t, err)
}

func TestSetFormProperties_NoProperties(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not make any request when no properties")
	})

	err := client.setFormProperties("123", map[string]interface{}{})
	require.NoError(t, err)
}

func TestAddQuestionsOneByOne_Success(t *testing.T) {
	var received []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		received = append(received, string(body))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := New("test-key")
	c.BaseURL = ts.URL

	questions := map[string]interface{}{
		"2": map[string]interface{}{"type": "control_email", "text": "Email"},
		"1": map[string]interface{}{"type": "control_head", "text": "Header"},
	}

	err := c.addQuestionsOneByOne("123", questions)
	require.NoError(t, err)
	assert.Len(t, received, 2)
}
