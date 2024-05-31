package tenablesc

import (
	"fmt"
)

const scanEndpoint = "/scan"

// Scan represents the request/response structure for https://docs.tenable.com/tenablesc/api/Scan.htm
type Scan struct {
	BaseInfo
	Assets               []BaseInfo          `json:"assets,omitempty"`
	ClassifyMitigatedAge string              `json:"classifyMitigatedAge,omitempty"`
	CreatedTime          UnixEpochStringTime `json:"createdTime,omitempty"`
	Creator              *UserInfo           `json:"creator,omitempty"`
	Credentials          []BaseInfo          `json:"credentials,omitempty"`
	DHCPTracking         FakeBool            `json:"dhcpTracking,omitempty"`
	EmailOnFinish        FakeBool            `json:"emailOnFinish,omitempty"`
	EmailOnLaunch        FakeBool            `json:"emailOnLaunch,omitempty"`
	InactivityTimeout    string              `json:"inactivityTimeout,omitempty"`
	IPList               string              `json:"ipList,omitempty"`
	MaxScanTime          string              `json:"maxScanTime,omitempty"`
	ModifiedTime         UnixEpochStringTime `json:"modifiedTime,omitempty"`
	NumDependents        ProbablyString      `json:"numDependents,omitempty"`
	Owner                *UserInfo           `json:"owner,omitempty"`
	OwnerGroup           *BaseInfo           `json:"ownerGroup,omitempty"`
	Plugin               *struct {
		ID ProbablyString `json:"id,omitempty"`
	} `json:"plugin,omitempty"`
	PluginID             ProbablyString `json:"pluginID,omitempty"`
	PluginPrefs          []string       `json:"pluginPrefs,omitempty"`
	Policy               *BaseInfo      `json:"policy,omitempty"`
	Repository           *BaseInfo      `json:"repository,omitempty"`
	Reports              []ScanReports  `json:"reports,omitempty"`
	RolloverType         string         `json:"rolloverType,omitempty"`
	Schedule             *ScanSchedule  `json:"schedule,omitempty"`
	ScanningVirtualHosts FakeBool       `json:"scanningVirtualHosts,omitempty"`
	Status               string         `json:"status,omitempty"`
	TimeoutAction        string         `json:"timeoutAction,omitempty"`
	Type                 string         `json:"type,omitempty"`
	Zone                 *BaseInfo      `json:"zone,omitempty"`
}

type ScanReports struct {
	ID           string `json:"id"`
	ReportSource string `json:"reportSource,omitempty"`
}

type ScanSchedule struct {
	ID          ProbablyString         `json:"id,omitempty"`
	DependentID string                 `json:"dependentID,omitempty"`
	Type        string                 `json:"type,omitempty"`
	Start       string                 `json:"start,omitempty"`
	RepeatRule  string                 `json:"repeatRule,omitempty"`
	Enabled     FakeBool               `json:"enabled,omitempty"`
	ObjectType  ProbablyString         `json:"objectType,omitempty"`
	NextRun     int                    `json:"nextRun,omitempty"`
	Dependent   *ScanDependentSchedule `json:"dependent,omitempty"`
}

type ScanDependentSchedule struct {
	BaseInfo
	Status string `json:"status,omitempty"`
}

type ScanStartResult struct {
	ScanID     string `json:"scanID"`
	ScanResult struct {
		BaseInfo
		InitiatorID    int    `json:"initiatorID"`
		OwnerID        string `json:"ownerID"`
		ScanID         string `json:"scanID"`
		ResultsSyncID  int    `json:"resultsSyncID"`
		JobID          string `json:"jobID"`
		RepositoryID   string `json:"repositoryID"`
		Details        string `json:"details"`
		Status         string `json:"status"`
		DownloadFormat string `json:"downloadFormat"`
		DataFormat     string `json:"dataFormat"`
		ResultType     string `json:"resultType"`
	} `json:"scanResult"`
}

type orgScanResponse struct {
	Manageable []*Scan `json:"manageable" tenable:"recurse"`
	Usable     []*Scan `json:"usable" tenable:"recurse"`
}

func (o orgScanResponse) orgScanResponseToExternal() []*Scan {
	var spOut []*Scan
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

func (c *Client) GetAllScans() ([]*Scan, error) {

	var orgScan orgScanResponse
	_, err := c.getResource(scanEndpoint, &orgScan)
	if err == nil {
		return orgScan.orgScanResponseToExternal(), nil
	}

	var s []*Scan
	if _, err := c.getResource(scanEndpoint, &s); err != nil {
		return nil, fmt.Errorf("failed to get scans: %w", err)
	}
	return s, nil
}

func (c *Client) CreateScan(s *Scan) (*Scan, error) {
	resp := &Scan{}

	if _, err := c.postResource(scanEndpoint, s, resp); err != nil {
		return nil, fmt.Errorf("failed to create scan: %w", err)
	}

	return resp, nil

}

func (c *Client) StartScan(id string) (*ScanStartResult, error) {
	resp := &ScanStartResult{}

	if _, err := c.postResource(fmt.Sprintf("%s/%s/launch", scanEndpoint, id), nil, resp); err != nil {
		return nil, fmt.Errorf("failed to start scan id %s: %w", id, err)
	}

	return resp, nil
}

func (c *Client) GetScan(id string) (*Scan, error) {
	resp := &Scan{}

	if _, err := c.getResource(fmt.Sprintf("%s/%s", scanEndpoint, id), resp); err != nil {
		return nil, fmt.Errorf("failed to get scan id %s: %w", id, err)
	}

	return resp, nil
}

func (c *Client) UpdateScan(s *Scan) (*Scan, error) {
	resp := &Scan{}

	if _, err := c.patchResourceWithID(scanEndpoint, s, resp); err != nil {
		return nil, fmt.Errorf("failed to update scan: %w", err)
	}

	return resp, nil
}

func (c *Client) DeleteScan(id string) error {

	if _, err := c.deleteResource(fmt.Sprintf("%s/%s", scanEndpoint, id), nil, nil); err != nil {
		return fmt.Errorf("unable to delete scan with id %s: %w", id, err)
	}

	return nil
}
