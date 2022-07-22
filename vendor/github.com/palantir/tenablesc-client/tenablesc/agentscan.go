package tenablesc

import (
	"fmt"
)

const agentScanEndpoint = "/agentScan"
const agentGroupsEndpoint = "/agentGroup/%s/remote"

// AgentScan is the fields renderable to and from the API for https://docs.tenable.com/tenablesc/api/Agent-Scan.htm
type AgentScan struct {
	BaseInfo
	Repository    BaseInfo       `json:"repository,omitempty"`
	NessusManager BaseInfo       `json:"nessusManager,omitempty"`
	ScanWindow    ProbablyString `json:"scanWindow,omitempty"`
	AgentGroups   []AgentGroup   `json:"agentGroups,omitempty"`
	Type          string         `json:"type,omitempty"`
	EmailOnLaunch FakeBool       `json:"emailOnLaunch,omitempty"`
	EmailOnFinish FakeBool       `json:"emailOnFinish,omitempty"`
	Schedule      *ScanSchedule  `json:"schedule,omitempty"`
}

type agentScanResponse struct {
	Manageable []*AgentScan `json:"manageable" tenable:"recurse"`
	Useable    []*AgentScan `json:"useable" tenable:"recurse"`
}

// AgentGroup is the fields describing an Agent Group reference in
// https://docs.tenable.com/tenablesc/api/Agent-Scan.htm ;
// for input, only the ID is needed.
type AgentGroup struct {
	BaseInfo
	NessusManagerID string `json:"nessusManagerID,omitempty"`
	RemoteID        int    `json:"remoteID,omitempty"`
}

func (c *Client) GetAllAgentScans() ([]*AgentScan, error) {
	resp := agentScanResponse{}

	if _, err := c.getResource(agentScanEndpoint, &resp); err != nil {
		return nil, fmt.Errorf("failed to get all agent scans: %w", err)
	}

	return resp.Manageable, nil
}

func (c *Client) GetAgentGroupsForScanner(id string) ([]*AgentGroup, error) {
	var resp []*AgentGroup

	if _, err := c.getResource(fmt.Sprintf(agentGroupsEndpoint, id), &resp); err != nil {
		return nil, fmt.Errorf("failed to get agent groups for scanner %s: %w", id, err)
	}

	return resp, nil
}

func (c *Client) CreateAgentScan(s *AgentScan) (*AgentScan, error) {
	resp := &AgentScan{}

	if _, err := c.postResource(agentScanEndpoint, s, resp); err != nil {
		return nil, fmt.Errorf("failed to create agent scan: %w", err)
	}

	return resp, nil
}

func (c *Client) UpdateAgentScan(s *AgentScan) (*AgentScan, error) {
	resp := &AgentScan{}

	if _, err := c.patchResourceWithID(agentScanEndpoint, s, resp); err != nil {
		return nil, fmt.Errorf("failed to update agent scan: %w", err)
	}

	return resp, nil
}

func (c *Client) DeleteAgentScan(id string) error {
	if _, err := c.deleteResource(fmt.Sprintf("%s/%s", agentScanEndpoint, id), nil, nil); err != nil {
		return fmt.Errorf("failed to delete agent scan %s: %w", id, err)
	}

	return nil
}
