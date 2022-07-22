package tenablesc

import (
	"fmt"
)

const scanPolicyTemplateEndpoint = "/policyTemplate"

// ScanPolicyTemplate represents the request/response structure from https://docs.tenable.com/tenablesc/api/Scan-Policy-Templates.htm
type ScanPolicyTemplate struct {
	BaseInfo
	Editor string `json:"editor,omitempty"`
	// DetailedEditor is input-only for unclear reasons.
	DetailedEditor     string              `json:"detailedEditor,omitempty"`
	CreatedTime        UnixEpochStringTime `json:"createdTime,omitempty"`
	ModifiedTime       UnixEpochStringTime `json:"modifiedTime,omitempty"`
	TemplatePubTime    UnixEpochStringTime `json:"templatePubTime,omitempty"`
	TemplateModTime    UnixEpochStringTime `json:"templateModTime,omitempty"`
	TemplateDefModTime UnixEpochStringTime `json:"templateDefModTime,omitempty"`
}

func (c *Client) GetAllScanPolicyTemplates() ([]*ScanPolicyTemplate, error) {
	var s []*ScanPolicyTemplate

	if _, err := c.getResource(scanPolicyTemplateEndpoint, &s); err != nil {
		return nil, fmt.Errorf("failed to get scan policy templates: %w", err)
	}

	return s, nil
}

func (c *Client) GetScanPolicyTemplate(id string) (*ScanPolicyTemplate, error) {
	resp := &ScanPolicyTemplate{}

	if _, err := c.getResource(fmt.Sprintf("%s/%s", scanPolicyTemplateEndpoint, id), resp); err != nil {
		return nil, fmt.Errorf("failed to get scanPolicyTemplate id %s: %w", id, err)
	}
	return resp, nil
}

func (c *Client) CreateScanPolicyTemplate(s *ScanPolicyTemplate) (*ScanPolicyTemplate, error) {
	resp := &ScanPolicyTemplate{}

	if _, err := c.postResource(scanPolicyTemplateEndpoint, s, resp); err != nil {
		return nil, fmt.Errorf("failed to create scanPolicyTemplate: %w", err)
	}
	return resp, nil
}

func (c *Client) DeleteScanPolicyTemplate(id string) error {

	if _, err := c.deleteResource(fmt.Sprintf("%s/%s", scanPolicyTemplateEndpoint, id), nil, nil); err != nil {
		return fmt.Errorf("unable to delete scan policy template with id %s: %w", id, err)
	}

	return nil
}
