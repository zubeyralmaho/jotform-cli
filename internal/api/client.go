package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const DefaultBaseURL = "https://api.jotform.com"

type Client struct {
	APIKey  string
	BaseURL string
	http    *http.Client
}

func New(apiKey string) *Client {
	return &Client{
		APIKey:  apiKey,
		BaseURL: DefaultBaseURL,
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

// apiResponse is the standard Jotform API envelope.
type apiResponse[T any] struct {
	ResponseCode int    `json:"responseCode"`
	Message      string `json:"message"`
	Content      T      `json:"content"`
}

func (c *Client) get(path string, out any) error {
	sep := "?"
	for _, ch := range path {
		if ch == '?' {
			sep = "&"
			break
		}
	}
	url := fmt.Sprintf("%s%s%sapiKey=%s", c.BaseURL, path, sep, c.APIKey)
	resp, err := c.http.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid API key — run `jotform auth login`")
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, body)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
