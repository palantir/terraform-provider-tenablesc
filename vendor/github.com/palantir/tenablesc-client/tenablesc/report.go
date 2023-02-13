package tenablesc

import (
	"fmt"
)

const reportEndpoint = "/report"

// Report represents request/response structure from https://docs.tenable.com/tenablesc/api/Report.htm
type Report struct {
	BaseInfo
	ReportDefinitionID string              `json:"reportDefinitionID"`
	JobID              string              `json:"jobID"`
	Type               string              `json:"type"`
	Status             string              `json:"status"`
	Running            string              `json:"running"`
	ErrorDetails       string              `json:"errorDetails"`
	TotalSteps         string              `json:"totalSteps"`
	CompletedSteps     string              `json:"completedSteps"`
	StartTime          UnixEpochStringTime `json:"startTime"`
	FinishTime         UnixEpochStringTime `json:"finishTime"`
	OwnerGID           string              `json:"ownerGID"`
	PubSites           []BaseInfo          `json:"pubSites"`
	Creator            UserInfo            `json:"creator"`
	Owner              UserInfo            `json:"owner"`
	OwnerGroup         BaseInfo            `json:"ownerGroup"`
}

type allReportsResponse struct {
	Manageable []*Report `json:"manageable" tenable:"recurse"`
	Usable     []*Report `json:"usable" tenable:"recurse"`
}

func (o allReportsResponse) allReportsToExternal() []*Report {
	var spOut []*Report
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

func (c *Client) GetAllReports() ([]*Report, error) {

	var allReports allReportsResponse
	if _, err := c.getResource(reportEndpoint, &allReports); err != nil {
		return nil, fmt.Errorf("failed to get reports: %w", err)
	}

	return allReports.allReportsToExternal(), nil
}

func (c *Client) GetReport(id string) (*Report, error) {
	resp := &Report{}

	if _, err := c.getResource(fmt.Sprintf("%s/%s", reportEndpoint, id), resp); err != nil {
		return nil, fmt.Errorf("failed to get report id %s: %w", id, err)
	}

	return resp, nil
}
