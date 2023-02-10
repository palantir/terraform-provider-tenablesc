package tenablesc

import (
	"fmt"
)

const reportDefinitionEndpoint = "/reportDefinition"

// ReportDefinitionBase represents the base request/response structure from https://docs.tenable.com/tenablesc/api/Report-Definition.htm
type ReportDefinitionBase struct {
	BaseInfo
}

type allReportDefinitionsResponse struct {
	Manageable []*ReportDefinitionBase `json:"manageable" tenable:"recurse"`
	Usable     []*ReportDefinitionBase `json:"usable" tenable:"recurse"`
}

func (o allReportDefinitionsResponse) allReportDefinitionsToExternal() []*ReportDefinitionBase {
	var spOut []*ReportDefinitionBase
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

func (c *Client) GetAllReportDefinitions() ([]*ReportDefinitionBase, error) {
	var allReportDefinitions *allReportDefinitionsResponse
	if _, err := c.getResource(reportDefinitionEndpoint, &allReportDefinitions); err != nil {
		return nil, fmt.Errorf("failed to get report definitions: %w", err)
	}

	return allReportDefinitions.allReportDefinitionsToExternal(), nil
}
