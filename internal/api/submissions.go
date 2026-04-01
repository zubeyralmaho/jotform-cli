package api

import "fmt"

type Submission struct {
	ID        string                 `json:"id"`
	FormID    string                 `json:"form_id"`
	CreatedAt string                 `json:"created_at"`
	Status    string                 `json:"status"`
	Answers   map[string]interface{} `json:"answers"`
}

func (c *Client) GetSubmissions(formID string, offset, limit int, orderBy, direction string) ([]Submission, error) {
	var resp apiResponse[[]Submission]
	path := fmt.Sprintf("/form/%s/submissions?offset=%d&limit=%d&orderby=%s&direction=%s",
		formID, offset, limit, orderBy, direction)
	if err := c.get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Content, nil
}
