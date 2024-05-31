package tenablesc

import (
	"fmt"
)

const currentUserEndpoint = "/currentUser"

// CurrentUser represents response structure for https://docs.tenable.com/tenablesc/api/Current-User.htm
//
//	This is perfect for testing credential validity and user type.
type CurrentUser struct {
	BaseInfo
	Status       string   `json:"status,omitempty"`
	Username     string   `json:"username,omitempty"`
	AuthType     string   `json:"authType,omitempty"`
	OrgName      string   `json:"orgName,omitempty"`
	Organization BaseInfo `json:"organization,omitempty"`
	Role         BaseInfo `json:"role,omitempty"`
	Group        BaseInfo `json:"group,omitempty"`
}

func (c *Client) GetCurrentUser() (*CurrentUser, error) {
	u := &CurrentUser{}

	if _, err := c.getResource(currentUserEndpoint, u); err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	return u, nil
}
