package tenablesc

import (
	"fmt"
)

// Organization represents the request and response structure for https://docs.tenable.com/tenablesc/api/Organization.htm
type Organization struct {
	BaseInfo
	Email   string `json:"email,omitempty"`
	Address string `json:"address,omitempty"`
	City    string `json:"city,omitempty"`
	State   string `json:"state,omitempty"`
	Phone   string `json:"phone,omitempty"`
	Fax     string `json:"fax,omitempty"`

	ZoneSelection string `json:"zoneSelection,omitempty"`
	RestrictedIPs string `json:"restrictedIPs,omitempty"`

	VulnScoreLow      string `json:"vulnScoreLow,omitempty"`
	VulnScoreMedium   string `json:"vulnScoreMedium,omitempty"`
	VulnScoreHigh     string `json:"vulnScoreHigh,omitempty"`
	VulnScoringSystem string `json:"vulnScoringSystem,omitempty"`

	CreatedTime  UnixEpochStringTime `json:"createdTime,omitempty"`
	ModifiedTime UnixEpochStringTime `json:"modifiedTime,omitempty"`

	UserCount string `json:"userCount,omitempty"`

	Repositories []OrganizationRepository `json:"repositories,omitempty"`
	Zones        []BaseInfo               `json:"zones,omitempty"`
}

type OrganizationRepository struct {
	RepoBaseFields
	GroupAssign string `json:"groupAssign,omitempty"`
}

const orgsEndpoint = "/organization"

func (c *Client) CreateOrganization(org *Organization) (*Organization, error) {
	out := &Organization{}

	if _, err := c.postResource(orgsEndpoint, org, out); err != nil {
		return nil, fmt.Errorf("failed to create org: %w", err)
	}

	return out, nil
}

func (c *Client) GetAllOrganizations() ([]*Organization, error) {
	var res []*Organization

	if _, err := c.getResource(orgsEndpoint, &res); err != nil {
		return res, fmt.Errorf("failed to get all orgs: %w", err)
	}

	return res, nil
}

func (c *Client) GetOrganization(id string) (*Organization, error) {
	res := &Organization{}

	if _, err := c.getResource(fmt.Sprintf("%s/%s", orgsEndpoint, id), &res); err != nil {
		return res, fmt.Errorf("failed to get org id %s: %w", id, err)
	}

	return res, nil
}

func (c *Client) UpdateOrganization(org *Organization) (*Organization, error) {
	out := &Organization{}

	if _, err := c.patchResourceWithID(orgsEndpoint, org, out); err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return out, nil
}

func (c *Client) DeleteOrganization(id string) error {
	if _, err := c.deleteResource(fmt.Sprintf("%s/%s", orgsEndpoint, id), nil, nil); err != nil {
		return fmt.Errorf("failed to delete org %s, %w", id, err)
	}
	return nil
}
