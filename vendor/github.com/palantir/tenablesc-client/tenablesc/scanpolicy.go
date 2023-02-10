package tenablesc

import (
	"fmt"
)

const scanPolicyEndpoint = "/policy"

// ScanPolicy represents request/response structure from https://docs.tenable.com/tenablesc/api/Scan-Policy.htm
type ScanPolicy struct {
	BaseInfo
	// to update a scan, you must pass an empty context. do not omitempty.
	Context      string              `json:"context"`
	Tags         string              `json:"tags,omitempty"`
	CreatedTime  UnixEpochStringTime `json:"createdTime,omitempty"`
	ModifiedTime UnixEpochStringTime `json:"modifiedTime,omitempty"`
	AuditFiles   []BaseInfo          `json:"auditFiles,omitempty"`
	// Somtimes a map and sometimes an array :(
	Preferences    interface{}          `json:"preferences,omitempty"`
	RemovePrefs    []string             `json:"removePrefs,omitempty"`
	PolicyTemplate *BaseInfo            `json:"policyTemplate,omitempty"`
	Owner          *UserInfo            `json:"owner,omitempty"`
	Creator        *UserInfo            `json:"creator,omitempty"`
	Families       []ScanPolicyFamilies `json:"families,omitempty"`
}

type ScanPolicyFamilies struct {
	ID      string     `json:"id"`
	Name    string     `json:"name,omitempty"`
	Count   string     `json:"count,omitempty"`
	State   string     `json:"state,omitempty"`
	Type    string     `json:"type,omitempty"`
	Plugins []BaseInfo `json:"plugins,omitempty"`
}

type orgScanPolicyResponse struct {
	Manageable []*ScanPolicy `json:"manageable" tenable:"recurse"`
	Usable     []*ScanPolicy `json:"usable" tenable:"recurse"`
}

func (o orgScanPolicyResponse) orgScanPolicyResponseToExternal() []*ScanPolicy {
	var spOut []*ScanPolicy
	spMap := make(map[ProbablyString]bool)

	for _, o := range o.Usable {
		spOut = append(spOut, o)
		spMap[o.ID] = true
	}
	for _, o := range o.Manageable {
		if _, exists := spMap[o.ID]; !exists {
			spOut = append(spOut, o)
			spMap[o.ID] = true
		}
	}

	return spOut
}

func (c *Client) GetAllScanPolicies() ([]*ScanPolicy, error) {

	var orgScanPolicy orgScanPolicyResponse
	_, err := c.getResource(scanPolicyEndpoint, &orgScanPolicy)
	if err == nil {
		return orgScanPolicy.orgScanPolicyResponseToExternal(), nil
	}
	// no? okay, ask again but assume we get a single struct.

	var s []*ScanPolicy

	if _, err := c.getResource(scanPolicyEndpoint, &s); err != nil {
		return nil, fmt.Errorf("failed to get scan policies: %w", err)
	}
	return s, nil
}

func (c *Client) CreateScanPolicy(s *ScanPolicy) (*ScanPolicy, error) {
	resp := &ScanPolicy{}

	if _, err := c.postResource(scanPolicyEndpoint, s, resp); err != nil {
		return nil, fmt.Errorf("failed to create scan policy: %w", err)
	}

	return resp, nil

}

func (c *Client) GetScanPolicy(id string) (*ScanPolicy, error) {
	resp := &ScanPolicy{}

	if _, err := c.getResource(fmt.Sprintf("%s/%s", scanPolicyEndpoint, id), resp); err != nil {
		return nil, fmt.Errorf("failed to get scan policy id %s: %w", id, err)
	}

	return resp, nil
}

func (c *Client) UpdateScanPolicy(s *ScanPolicy) (*ScanPolicy, error) {
	resp := &ScanPolicy{}

	if _, err := c.patchResourceWithID(scanPolicyEndpoint, s, resp); err != nil {
		return nil, fmt.Errorf("failed to update scan policy: %w", err)
	}

	return resp, nil
}

func (c *Client) DeleteScanPolicy(id string) error {

	if _, err := c.deleteResource(fmt.Sprintf("%s/%s", scanPolicyEndpoint, id), nil, nil); err != nil {
		return fmt.Errorf("unable to delete scan policy with id %s: %w", id, err)
	}

	return nil
}
