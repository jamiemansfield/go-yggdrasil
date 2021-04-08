package yggdrasil

import (
	"net/http"
)

var (
	AgentMinecraft = Agent{
		Name:    "Minecraft",
		Version: 1,
	}
	AgentScrolls = Agent{
		Name:    "Scrolls",
		Version: 1,
	}
)

type Agent struct {
	Name string `json:"name"`
	Version int `json:"version"`
}

// Authenticate attempts to authenticate with the Yggdrasil authentication
// server, using the given credentials.
//
// The clientToken needs to be specified if you wish to refresh the access
// token in future.
//
// User data will always be requested from this API call.
func (c *Client) Authenticate(agent Agent, username string, password string, clientToken string) (*AuthenticateResponse, error) {
	authRequest := authenticateRequest{
		Agent:       agent,
		Username:    username,
		Password:    password,
		ClientToken: clientToken,
		RequestUser: true,
	}

	req, err := c.NewRequest(http.MethodPost, "authenticate", authRequest)
	if err != nil {
		return nil, err
	}

	var response AuthenticateResponse
	if _, err := c.Do(req, &response); err != nil {
		return nil, err
	}

	return &response, err
}

type authenticateRequest struct {
	Agent Agent `json:"agent"`
	Username string `json:"username"`
	Password string `json:"password"`
	ClientToken string `json:"clientToken"`
	RequestUser bool `json:"requestUser"`
}

type AuthenticateResponse struct {
	AccessToken string `json:"accessToken"`
	ClientToken string `json:"clientToken"`
}
