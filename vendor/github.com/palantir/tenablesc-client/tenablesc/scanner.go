package tenablesc

import (
	"fmt"
)

const (
	scannerEndpoint    = "/scanner"
	updateScanEndpoint = "/updateStatus"
)

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

type UpdateScannersStatus struct {
	Status []struct {
		BaseInfo
		Status string `json:"status"`
	} `json:"status"`
}

func (c *Client) UpdateScanners() (*UpdateScannersStatus, error) {
	var resp *UpdateScannersStatus

	if _, err := c.postResource(
		fmt.Sprintf("%s%s", scannerEndpoint, updateScanEndpoint),
		nil,
		&resp,
	); err != nil {
		return nil, fmt.Errorf("could not update scanner status: %w", err)
	}

	return resp, nil
}
