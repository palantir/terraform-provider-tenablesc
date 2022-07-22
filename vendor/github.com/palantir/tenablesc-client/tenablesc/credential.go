package tenablesc

import (
	"fmt"
)

const credentialEndpoint = "/credential"

// Credential is massively pared back from the possible types in https://docs.tenable.com/tenablesc/api/Credential.htm
// This is not wired to directly manage credentials, only to find them.
type Credential struct {
	BaseInfo
	Type      string   `json:"type"`
	CanUse    FakeBool `json:"canUse,omitempty"`
	CanManage FakeBool `json:"canManage,omitempty"`
	Tags      string   `json:"tags,omitempty"`
}

type allCredentialsInternal struct {
	Usable     []*Credential `json:"usable,omitempty"`
	Manageable []*Credential `json:"manageable,omitempty"`
}

func (c *Client) GetAllCredentials() ([]*Credential, error) {
	var creds []*Credential

	allCredentials := &allCredentialsInternal{}

	if _, err := c.getResource(credentialEndpoint, &allCredentials); err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	credMap := make(map[ProbablyString]bool)

	for _, c := range allCredentials.Usable {
		credMap[c.ID] = true
		creds = append(creds, c)
	}
	for _, c := range allCredentials.Manageable {
		if _, exists := credMap[c.ID]; !exists {
			creds = append(creds, c)
			credMap[c.ID] = true
		}
	}

	return creds, nil
}

func (c *Client) GetCredential(id string) (*Credential, error) {
	resp := &Credential{}

	if _, err := c.getResource(fmt.Sprintf("%s/%s", credentialEndpoint, id), resp); err != nil {
		return nil, fmt.Errorf("failed to get credential id %s: %w", id, err)
	}

	return resp, nil
}
