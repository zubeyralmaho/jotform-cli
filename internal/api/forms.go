package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
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

	form := resp.Content
	if form.Questions == nil {
		form.Questions = map[string]interface{}{}
	}
	if form.Properties == nil {
		form.Properties = map[string]interface{}{}
	}

	// Enrich with dedicated endpoints. If one endpoint fails, keep base form data.
	var qResp apiResponse[map[string]interface{}]
	if err := c.get("/form/"+id+"/questions", &qResp); err == nil && qResp.Content != nil {
		form.Questions = qResp.Content
	}

	var pResp apiResponse[map[string]interface{}]
	if err := c.get("/form/"+id+"/properties", &pResp); err == nil && pResp.Content != nil {
		form.Properties = pResp.Content
		if form.Title == "" {
			if title, ok := pResp.Content["title"].(string); ok {
				form.Title = title
			}
		}
	}

	return &form, nil
}

func (c *Client) CreateForm(schema map[string]interface{}) (*Form, error) {
	values := flattenFormSchema(schema)
	body := strings.NewReader(values.Encode())

	url := fmt.Sprintf("%s/user/forms?apiKey=%s", c.BaseURL, c.APIKey)
	resp, err := c.http.Post(url, "application/x-www-form-urlencoded", body)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

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

	// Some Jotform accounts/endpoints create the shell form first.
	// Apply full schema explicitly to ensure questions/properties are persisted.
	if err := c.setFormProperties(apiResp.Content.ID, schema); err != nil {
		return nil, err
	}
	if err := c.setFormQuestions(apiResp.Content.ID, schema); err != nil {
		return nil, err
	}

	return &apiResp.Content, nil
}

func (c *Client) UpdateForm(id string, schema map[string]interface{}) (*Form, error) {
	if err := c.setFormProperties(id, schema); err != nil {
		return nil, err
	}
	if err := c.setFormQuestions(id, schema); err != nil {
		return nil, err
	}

	fullForm, err := c.GetForm(id)
	if err != nil {
		return nil, err
	}

	return &Form{
		ID:    fullForm.ID,
		Title: fullForm.Title,
		URL:   fmt.Sprintf("https://form.jotform.com/%s", fullForm.ID),
	}, nil
}

func (c *Client) setFormProperties(id string, schema map[string]interface{}) error {
	properties := url.Values{}

	if rawProps, ok := schema["properties"]; ok {
		if props, ok := rawProps.(map[string]interface{}); ok {
			for k, v := range props {
				properties.Set("properties["+k+"]", fmt.Sprint(v))
			}
		}
	}

	if title, ok := schema["title"]; ok && fmt.Sprint(title) != "" {
		properties.Set("properties[title]", fmt.Sprint(title))
	}

	if len(properties) == 0 {
		return nil
	}

	req, err := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/form/%s/properties?apiKey=%s", c.BaseURL, id, c.APIKey),
		strings.NewReader(properties.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (c *Client) setFormQuestions(id string, schema map[string]interface{}) error {
	rawQuestions, ok := schema["questions"]
	if !ok {
		return nil
	}

	questions, ok := rawQuestions.(map[string]interface{})
	if !ok {
		return nil
	}

	payload := map[string]interface{}{"questions": questions}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut,
		fmt.Sprintf("%s/form/%s/questions?apiKey=%s", c.BaseURL, id, c.APIKey),
		strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		if err := c.addQuestionsOneByOne(id, questions); err != nil {
			return fmt.Errorf("API error %d: %s; fallback failed: %w", resp.StatusCode, string(respBody), err)
		}
		return nil
	}

	return nil
}

func (c *Client) addQuestionsOneByOne(id string, questions map[string]interface{}) error {
	orders := make([]string, 0, len(questions))
	for order := range questions {
		orders = append(orders, order)
	}
	sort.Slice(orders, func(i, j int) bool {
		ii, errI := strconv.Atoi(orders[i])
		jj, errJ := strconv.Atoi(orders[j])
		if errI == nil && errJ == nil {
			return ii < jj
		}
		return orders[i] < orders[j]
	})

	for _, order := range orders {
		qRaw := questions[order]
		q, ok := qRaw.(map[string]interface{})
		if !ok {
			continue
		}

		values := url.Values{}
		for k, v := range q {
			values.Set("question["+k+"]", fmt.Sprint(v))
		}

		req, err := http.NewRequest(http.MethodPost,
			fmt.Sprintf("%s/form/%s/questions?apiKey=%s", c.BaseURL, id, c.APIKey),
			strings.NewReader(values.Encode()))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode >= 400 {
			respBody, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			return fmt.Errorf("API error %d while adding question %s: %s", resp.StatusCode, order, string(respBody))
		}
		_ = resp.Body.Close()
	}

	return nil
}

func flattenFormSchema(schema map[string]interface{}) url.Values {
	values := url.Values{}

	if rawProps, ok := schema["properties"]; ok {
		if props, ok := rawProps.(map[string]interface{}); ok {
			for k, v := range props {
				values.Set("properties["+k+"]", fmt.Sprint(v))
			}
		}
	}

	if title, ok := schema["title"]; ok && fmt.Sprint(title) != "" {
		values.Set("properties[title]", fmt.Sprint(title))
	}

	if rawQuestions, ok := schema["questions"]; ok {
		if questions, ok := rawQuestions.(map[string]interface{}); ok {
			for order, qRaw := range questions {
				q, ok := qRaw.(map[string]interface{})
				if !ok {
					continue
				}
				for k, v := range q {
					values.Set("questions["+order+"]["+k+"]", fmt.Sprint(v))
				}
			}
		}
	}

	return values
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
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("delete failed: %d", resp.StatusCode)
	}
	return nil
}
