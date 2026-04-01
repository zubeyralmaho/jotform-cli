package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Form struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	Status  string `json:"status"`
	Created string `json:"created_at"`
	Updated string `json:"updated_at"`
	Count   Count  `json:"count"`
}

// Count accepts both numeric and string count values returned by the API.
type Count string

func (c *Count) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*c = ""
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*c = Count(s)
		return nil
	}

	var n int64
	if err := json.Unmarshal(data, &n); err == nil {
		*c = Count(strconv.FormatInt(n, 10))
		return nil
	}

	return fmt.Errorf("invalid count value: %s", string(data))
}

type FormProperties struct {
	ID         string                 `json:"id"`
	Title      string                 `json:"title"`
	Questions  map[string]interface{} `json:"questions"`
	Properties map[string]interface{} `json:"properties"`
}

func (c *Client) ListForms(offset, limit int) ([]Form, error) {
	var resp apiResponse[[]Form]
	path := fmt.Sprintf("/user/forms?offset=%d&limit=%d", offset, limit)
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Content, nil
}

func (c *Client) GetForm(id string) (*FormProperties, error) {
	var resp apiResponse[FormProperties]
	if err := c.get("/form/"+id, &resp); err != nil {
		return nil, err
	}
	return &resp.Content, nil
}

func (c *Client) CreateForm(schema map[string]interface{}) (*Form, error) {
	body, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/user/forms?apiKey=%s", c.BaseURL, c.APIKey)
	resp, err := c.http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("API error 401: unauthorized. Check API key permissions for form creation. Details: %s", string(body))
		}
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var apiResp apiResponse[Form]
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	return &apiResp.Content, nil
}

func (c *Client) DeleteForm(id string) error {
	req, err := http.NewRequest(http.MethodDelete,
		fmt.Sprintf("%s/form/%s?apiKey=%s", c.BaseURL, id, c.APIKey), nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("delete failed: %d", resp.StatusCode)
	}
	return nil
}
