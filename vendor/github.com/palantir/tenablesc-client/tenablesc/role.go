package tenablesc

import (
	"fmt"
)

const roleEndpoint = "/role"

// Role represents request/response structure for https://docs.tenable.com/tenablesc/api/Role.htm
type Role struct {
	BaseInfo
	PermManageGroups             FakeBool `json:"permManageGroups,omitempty"`
	PermManageRoles              FakeBool `json:"permManageRoles,omitempty"`
	PermManageImages             FakeBool `json:"permManageImages,omitempty"`
	PermManageGroupRelationships FakeBool `json:"permManageGroupRelationships,omitempty"`
	PermManageBlackoutWindows    FakeBool `json:"permManageBlackoutWindows,omitempty"`
	PermManageAttributeSets      FakeBool `json:"permManageAttributeSets,omitempty"`
	PermCreateTickets            FakeBool `json:"permCreateTickets,omitempty"`
	PermCreateAlerts             FakeBool `json:"permCreateAlerts,omitempty"`
	PermCreateAuditFiles         FakeBool `json:"permCreateAuditFiles,omitempty"`
	PermCreateLDAPAssets         FakeBool `json:"permCreateLDAPAssets,omitempty"`
	PermCreatePolicies           FakeBool `json:"permCreatePolicies,omitempty"`
	PermPurgeTickets             FakeBool `json:"permPurgeTickets,omitempty"`
	PermPurgeScanResults         FakeBool `json:"permPurgeScanResults,omitempty"`
	PermPurgeReportResults       FakeBool `json:"permPurgeReportResults,omitempty"`
	PermAgentsScan               FakeBool `json:"permAgentsScan,omitempty"`
	PermShareObjects             FakeBool `json:"permShareObjects,omitempty"`
	PermUpdateFeeds              FakeBool `json:"permUpdateFeeds,omitempty"`
	PermUploadNessusResults      FakeBool `json:"permUploadNessusResults,omitempty"`
	PermViewOrgLogs              FakeBool `json:"permViewOrgLogs,omitempty"`
	PermManageAcceptRiskRules    FakeBool `json:"permManageAcceptRiskRules,omitempty"`
	PermManageRecastRiskRules    FakeBool `json:"permManageRecastRiskRules,omitempty"`

	// PermScan is treated as bool here, but it expects the API value to be 'full' (true) or 'none' (false).
	//   This quirk is marshelled/unmarshalled correctly internally.
	PermScan FakeBool `json:"permScan,omitempty"`
}

const (
	permScanTrue  = "full"
	permScanFalse = "none"
)

// For API use only.
func (r *Role) toInternal() *Role {

	role := *r

	if r.PermScan.AsBool() {
		role.PermScan = permScanTrue
	} else {
		role.PermScan = permScanFalse
	}

	return &role
}

// for general consumption only.
func (r *Role) toExternal() *Role {

	role := *r

	if r.PermScan == permScanTrue {
		role.PermScan = FakeTrue
	} else {
		role.PermScan = FakeFalse
	}

	return &role
}

func roleSliceToExternal(rsi []*Role) []*Role {
	rse := make([]*Role, 0, len(rsi))

	for _, r := range rsi {
		re := r.toExternal()
		rse = append(rse, re)
	}
	return rse
}

func (c *Client) GetAllRoles() ([]*Role, error) {

	var r []*Role

	if _, err := c.getResource(roleEndpoint, &r); err != nil {
		return nil, fmt.Errorf("could not get roles: %w", err)
	}

	return roleSliceToExternal(r), nil
}

func (c *Client) CreateRole(role *Role) (*Role, error) {
	resp := &Role{}

	if _, err := c.postResource(roleEndpoint, role.toInternal(), resp); err != nil {
		return nil, fmt.Errorf("could not create role: %w", err)
	}
	return resp.toExternal(), nil
}

func (c *Client) GetRole(id string) (*Role, error) {
	resp := &Role{}

	if _, err := c.getResource(fmt.Sprintf("%s/%s", roleEndpoint, id), resp); err != nil {
		return nil, fmt.Errorf("could not get role %s: %w", id, err)
	}

	return resp.toExternal(), nil
}

func (c *Client) UpdateRole(role *Role) (*Role, error) {
	resp := &Role{}

	if _, err := c.patchResourceWithID(roleEndpoint, role.toInternal(), resp); err != nil {
		return nil, fmt.Errorf("could not create role: %w", err)
	}
	return resp.toExternal(), nil
}

func (c *Client) DeleteRole(id string) error {
	if _, err := c.deleteResource(fmt.Sprintf("%s/%s", roleEndpoint, id), nil, nil); err != nil {
		return fmt.Errorf("failed to delete role %s, %w", id, err)
	}
	return nil
}
