package tenablesc

import (
	"fmt"
)

const scannerEndpoint = "/scanner"

type Scanner struct {
	BaseInfo
	Status       string   `json:"status"`
	AgentCapable FakeBool `json:"agentCapable"`
}

func (c *Client) GetAllScanners() ([]*Scanner, error) {
	var resp []*Scanner

	if _, err := c.getResource(scannerEndpoint, &resp); err != nil {
		return nil, fmt.Errorf("could not get scanners: %w", err)
	}

	return resp, nil
}
