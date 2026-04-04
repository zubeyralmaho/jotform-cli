package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Webhook represents a Jotform webhook configuration.
type Webhook struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// GetFormWebhooks returns all webhooks configured for a form.
func (c *Client) GetFormWebhooks(formID string) ([]Webhook, error) {
	var resp apiResponse[map[string]string]
	if err := c.get("/form/"+formID+"/webhooks", &resp); err != nil {
		return nil, err
	}

	var webhooks []Webhook
	for id, webhookURL := range resp.Content {
		webhooks = append(webhooks, Webhook{ID: id, URL: webhookURL})
	}
	return webhooks, nil
}

// CreateFormWebhook adds a webhook URL to a form.
func (c *Client) CreateFormWebhook(formID, webhookURL string) error {
	values := url.Values{}
	values.Set("webhookURL", webhookURL)

	req, err := http.NewRequest(http.MethodPost,
		fmt.Sprintf("%s/form/%s/webhooks?apiKey=%s", c.BaseURL, formID, c.APIKey),
		strings.NewReader(values.Encode()))
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
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// DeleteFormWebhook removes a webhook from a form by its ID.
func (c *Client) DeleteFormWebhook(formID, webhookID string) error {
	req, err := http.NewRequest(http.MethodDelete,
		fmt.Sprintf("%s/form/%s/webhooks/%s?apiKey=%s", c.BaseURL, formID, webhookID, c.APIKey), nil)
	if err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// TestWebhook sends a test POST request to a webhook URL to verify it's reachable.
func TestWebhook(webhookURL string) (int, error) {
	payload := `{"test": true, "source": "jotform-cli", "message": "This is a test webhook from jotform-cli"}`
	resp, err := http.Post(webhookURL, "application/json", strings.NewReader(payload))
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	// Read and discard the body
	_, _ = io.ReadAll(resp.Body)

	return resp.StatusCode, nil
}

// WebhookPayload represents the structure sent to webhook test endpoints.
type WebhookPayload struct {
	Test    bool   `json:"test"`
	Source  string `json:"source"`
	Message string `json:"message"`
}

// MarshalTestPayload returns the test payload as JSON bytes.
func MarshalTestPayload() ([]byte, error) {
	return json.Marshal(WebhookPayload{
		Test:    true,
		Source:  "jotform-cli",
		Message: "This is a test webhook from jotform-cli",
	})
}
