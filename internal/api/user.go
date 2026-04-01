package api

type User struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	AccountType string `json:"account_type"`
}

func (c *Client) GetUser() (*User, error) {
	var resp apiResponse[User]
	if err := c.get("/user", &resp); err != nil {
		return nil, err
	}
	return &resp.Content, nil
}
