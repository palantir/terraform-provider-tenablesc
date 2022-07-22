package tenablesc

import (
	"fmt"
)

const auditFilesEndpoint = "/auditFile"

// AuditFile respresents the request/response structure for https://docs.tenable.com/tenablesc/api/AuditFile.htm
type AuditFile struct {
	BaseInfo
	Type              string              `json:"type,omitempty"`
	Version           string              `json:"version,omitempty"`
	Status            string              `json:"status,omitempty"`
	Filename          string              `json:"filename,omitempty"`
	OriginalFilename  string              `json:"originalFilename,omitempty"`
	Variables         []AuditVariable     `json:"variables,omitempty"`
	CreatedTime       UnixEpochStringTime `json:"createdTime,omitempty"`
	ModifiedTime      UnixEpochStringTime `json:"modifiedTime,omitempty"`
	LastRefreshedTime UnixEpochStringTime `json:"lastRefreshedTime,omitempty"`
}

type AuditVariable struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type auditFileInternal struct {
	AuditFile
	TypeFields struct {
		Variables []AuditVariable `json:"variables,omitempty"`
	} `json:"typeFields,omitempty"`
}

func (a auditFileInternal) toExternal() *AuditFile {
	a.AuditFile.Variables = a.TypeFields.Variables

	return &a.AuditFile
}

func (c *Client) CreateAuditFile(a *AuditFile) (*AuditFile, error) {
	out := &auditFileInternal{}

	if _, err := c.postResource(auditFilesEndpoint, a, out); err != nil {
		return nil, fmt.Errorf("failed to create audit file: %w", err)
	}

	return out.toExternal(), nil
}

func (c *Client) GetAuditFile(id string) (*AuditFile, error) {
	res := &auditFileInternal{}

	if _, err := c.getResource(fmt.Sprintf("%s/%s", auditFilesEndpoint, id), &res); err != nil {
		return nil, fmt.Errorf("failed to get audit file id %s: %w", id, err)
	}

	return res.toExternal(), nil
}

func (c *Client) UpdateAuditFile(a *AuditFile) (*AuditFile, error) {
	out := &auditFileInternal{}

	if _, err := c.patchResourceWithID(auditFilesEndpoint, a, out); err != nil {
		return nil, fmt.Errorf("failed to update audit file: %w", err)
	}

	return out.toExternal(), nil
}

func (c *Client) DeleteAuditFile(id string) error {
	if _, err := c.deleteResource(fmt.Sprintf("%s/%s", auditFilesEndpoint, id), nil, nil); err != nil {
		return fmt.Errorf("failed to delete audit file %s, %w", id, err)
	}

	return nil
}
