package tenablesc

import (
	"fmt"
	"strings"
)

const scanZoneEndpoint = "/zone"

// ScanZone represents the request/response fields for https://docs.tenable.com/tenablesc/api/Scan-Zone.htm
type ScanZone struct {
	ScanZoneBaseFields
	// IPList is composed internally as a comma-separated list; we split and join for your convenience.
	IPList []string `json:"ipList,omitempty"`
}

type ScanZoneBaseFields struct {
	BaseInfo
	CreatedTime   string            `json:"createdTime,omitempty"`
	ModifiedTime  string            `json:"modifiedTime,omitempty"`
	Organizations []BaseInfo        `json:"organizations,omitempty"`
	Scanners      []ScanZoneScanner `json:"scanners,omitempty"`
}

type ScanZoneScanner struct {
	BaseInfo
	Status string `json:"status,omitempty"`
}

type scanZoneInternal struct {
	ScanZoneBaseFields
	IPList string `json:"ipList,omitempty"`
}

func (sz scanZoneInternal) toExternal() *ScanZone {

	scanZone := &ScanZone{
		ScanZoneBaseFields: sz.ScanZoneBaseFields,
	}
	scanZone.IPList = strings.Split(sz.IPList, ",")

	return scanZone
}

func (sz ScanZone) toInternal() *scanZoneInternal {
	scanZone := &scanZoneInternal{
		ScanZoneBaseFields: sz.ScanZoneBaseFields,
	}

	scanZone.IPList = strings.Join(sz.IPList, ",")

	return scanZone

}

func scanZoneSliceToExternal(szi []*scanZoneInternal) []*ScanZone {
	szs := make([]*ScanZone, 0, len(szi))

	for _, sz := range szi {
		sze := sz.toExternal()
		szs = append(szs, sze)
	}
	return szs
}

func (c *Client) GetAllScanZones() ([]*ScanZone, error) {
	var sz []*scanZoneInternal

	if _, err := c.getResource(scanZoneEndpoint, &sz); err != nil {
		return nil, fmt.Errorf("could not get scan zones: %w", err)
	}

	return scanZoneSliceToExternal(sz), nil
}

func (c *Client) CreateScanZone(sz *ScanZone) (*ScanZone, error) {
	resp := &scanZoneInternal{}
	if _, err := c.postResource(scanZoneEndpoint, sz.toInternal(), resp); err != nil {
		return nil, fmt.Errorf("could not create scan zone: %w", err)
	}
	return resp.toExternal(), nil
}

func (c *Client) GetScanZone(id string) (*ScanZone, error) {
	resp := &scanZoneInternal{}
	if _, err := c.getResource(fmt.Sprintf("%s/%s", scanZoneEndpoint, id), resp); err != nil {
		return nil, fmt.Errorf("could not get scan zone id %s: %w", id, err)
	}
	return resp.toExternal(), nil
}
func (c *Client) UpdateScanZone(sz *ScanZone) (*ScanZone, error) {
	resp := &scanZoneInternal{}

	if _, err := c.patchResourceWithID(scanZoneEndpoint, sz.toInternal(), resp); err != nil {
		return nil, fmt.Errorf("failed to update scan zone: %w", err)
	}

	return resp.toExternal(), nil
}

func (c *Client) DeleteScanZone(id string) error {
	if _, err := c.deleteResource(fmt.Sprintf("%s/%s", scanZoneEndpoint, id), nil, nil); err != nil {
		return fmt.Errorf("failed to delete scan zone %s, %w", id, err)
	}
	return nil
}
