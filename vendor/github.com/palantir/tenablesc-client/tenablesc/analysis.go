package tenablesc

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

const (
	analysisEndpoint = "/analysis"
)

// Analysis represents the fields used for requests against https://docs.tenable.com/tenablesc/api/Analysis.htm
// Note some fields are used only for certain combinations of type and sourcetype.
type Analysis struct {
	Type          string        `json:"type,omitempty"`
	Query         AnalysisQuery `json:"query,omitempty"`
	SourceType    string        `json:"sourceType,omitempty"`
	SortField     string        `json:"sortField,omitempty"`
	SortDirection string        `json:"sortDir,omitempty"`
	Columns       []BaseInfo    `json:"columns"` // note: this wants column names, not ids.
	// StartOffset is for type vuln only
	StartOffset string `json:"startOffset,omitempty"`
	// EndOffset is for type vuln only
	EndOffset string `json:"endOffset,omitempty"`
	// ScanID is for type vuln , sourcetype individual only
	ScanID string `json:"scanID,omitempty"`
	// View is for type vuln , sourcetype individual only
	View string `json:"view,omitempty"`
}

// AnalysisQuery represents the fields available for filtering queries.
// Field values are probably best discovered through building queries
// in browser developer tools and then interrogating the request payloads.
type AnalysisQuery struct {
	ID          string           `json:"id,omitempty"`
	Name        string           `json:"name,omitempty"`
	Description string           `json:"description,omitempty"`
	Context     string           `json:"context,omitempty"`
	Type        string           `json:"type,omitempty"`
	SourceType  string           `json:"sourceType,omitempty"`
	Status      string           `json:"status,omitempty"`
	Tool        string           `json:"tool,omitempty"`
	Filters     []AnalysisFilter `json:"filters,omitempty"`
}

// AnalysisFilter is the structure used for Analysis query filtering
type AnalysisFilter struct {
	FilterName string `json:"filterName"`
	Operator   string `json:"operator"`
	// Value is sometimes a string, sometimes a number, sometimes a BaseInfo (with ID only generally)
	Value interface{} `json:"value"`
}

// AnalysisResponseContainer is the output structure produced by an Analyze query.
// Results are further unmarshalled before return from call.
type AnalysisResponseContainer struct {
	TotalRecords             string          `json:"totalRecords,omitempty"`
	ReturnedRecords          int             `json:"returnedRecords,omitempty"`
	StartOffset              string          `json:"startOffset,omitempty"`
	EndOffset                string          `json:"endOffset,omitempty"`
	MatchingDataElementCount string          `json:"matchingDataElementCount,omitempty"`
	Results                  json.RawMessage `json:"results,omitempty"`
}

// VulnSumIPResult contains the structure used by the 'sumip' analysis tool.
type VulnSumIPResult struct {
	BiosGUID         string         `json:"biosGUID"`
	DNSName          string         `json:"dnsName"`
	IP               string         `json:"ip"`
	LastAuthRun      string         `json:"lastAuthRun"`
	LastUnauthRun    string         `json:"lastUnauthRun"`
	MacAddress       string         `json:"macAddress"`
	McafeeGUID       string         `json:"mcafeeGUID"`
	NetBiosName      string         `json:"netbiosName"`
	OSCPE            string         `json:"osCPE"`
	PluginSet        string         `json:"pluginSet"`
	PolicyName       string         `json:"policyName"`
	Repository       VulnRepository `json:"repository"`
	Score            string         `json:"score"`
	SeverityCritical string         `json:"severityCritical"`
	SeverityHigh     string         `json:"severityHigh"`
	SeverityInfo     string         `json:"severityInfo"`
	SeverityMedium   string         `json:"severityMedium"`
	SeverityLow      string         `json:"severityLow"`
	Total            string         `json:"total"`
	TPMID            string         `json:"tpmID"`
	Uniqueness       string         `json:"uniqueness"`
	UUID             string         `json:"uuid"`
}

// VulnSumDNSNameResult contains the structure used by the 'sumdnsname' analysis tool
type VulnSumDNSNameResult struct {
	DNSName          string         `json:"dnsName"`
	Repository       VulnRepository `json:"repository"`
	Score            string         `json:"score"`
	SeverityCritical string         `json:"severityCritical"`
	SeverityHigh     string         `json:"severityHigh"`
	SeverityInfo     string         `json:"severityInfo"`
	SeverityMedium   string         `json:"severityMedium"`
	SeverityLow      string         `json:"severityLow"`
	Total            string         `json:"total"`
}

// VulnIPSummaryResult contains the structure used by the 'vulnipsummary' analysis tool
type VulnIPSummaryResult struct {
	Name              string     `json:"name"`
	Family            VulnFamily `json:"family"`
	Hosts             VulnHosts  `json:"hosts"`
	PluginDescription string     `json:"pluginDescription"`
	PluginID          string     `json:"pluginID"`
	RepositoryID      string     `json:"repositoryID"`
	Severity          BaseInfo   `json:"severity"`
	Total             string     `json:"total"`
}

// VulnDetailsResult contains the structure used by the 'vulndetails' analysis tool
type VulnDetailsResult struct {
	AcceptRisk          string         `json:"acceptRisk"`
	BaseScore           string         `json:"baseScore"`
	BID                 string         `json:"bid"`
	CheckType           string         `json:"checkType"`
	CPE                 string         `json:"cpe"`
	CVE                 string         `json:"cve"`
	CVSSV3BaseScore     string         `json:"cvssV3BaseScore"`
	CVSSV3TemporalScore string         `json:"cvssV3TemporalScore"`
	CVSSV3Vector        string         `json:"cvssV3Vector"`
	CVSSVector          string         `json:"cvssVector"`
	Description         string         `json:"description"`
	DNSName             string         `json:"dnsName"`
	ExploitAvailable    string         `json:"exploitAvailable"`
	ExploitEase         string         `json:"exploitEase"`
	ExploitFrameworks   string         `json:"exploitFrameworks"`
	Family              VulnFamily     `json:"family"`
	FirstSeen           string         `json:"firstSeen"`
	HasBeenMitigated    string         `json:"hasBeenMitigated"`
	HostUniqueness      string         `json:"hostUniqueness"`
	IP                  string         `json:"ip"`
	IPS                 string         `json:"ips"`
	LastSeen            string         `json:"lastSeen"`
	MacAddress          string         `json:"macAddress"`
	NetbiosName         string         `json:"netbiosName"`
	PatchPubDate        string         `json:"patchPubDate"`
	PluginID            string         `json:"pluginID"`
	PluginInfo          string         `json:"pluginInfo"`
	PluginModDate       string         `json:"pluginModDate"`
	PluginName          string         `json:"pluginName"`
	PluginPubDate       string         `json:"pluginPubDate"`
	PluginText          string         `json:"pluginText"`
	Port                string         `json:"port"`
	Protocol            string         `json:"protocol"`
	RecastRisk          string         `json:"recastRisk"`
	Repository          VulnRepository `json:"repository"`
	RiskFactor          string         `json:"riskFactor"`
	SeeAlso             string         `json:"seeAlso"`
	Severity            BaseInfo       `json:"severity"`
	Solution            string         `json:"solution"`
	StigSeverity        string         `json:"stigSeverity"`
	Synopsis            string         `json:"synopsis"`
	TemporalScore       string         `json:"temporalScore"`
	UUID                string         `json:"uuid"`
	Version             string         `json:"version"`
	VPRContext          string         `json:"vprContext"`
	VPRScore            string         `json:"vprScore"`
	VulnPubDate         string         `json:"vulnPubDate"`
	XREF                string         `json:"xref"`
}

// VulnFamily information for a vulnerability
type VulnFamily struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// VulnHosts lists the hosts affected by the vulnerability
type VulnHosts struct {
	IPList      string     `json:"iplist"`
	Repository  Repository `json:"repository"`
	UUIDIPsList string     `json:"uuidIPsList"`
}

// VulnRepository includes information about the respository that the vulnerability came from
type VulnRepository struct {
	DataFormat  string `json:"dataFormat"`
	Description string `json:"description"`
	ID          string `json:"id"`
	Name        string `json:"name"`
}

// Analyze takes an arbitrary Analysis query, determines the expected container object, and writes to
// the resultsContainer expecting it to be the correct type.
// The partially-unmarshalled response is returned as well for metadata purposes.
func (c *Client) Analyze(a *Analysis, resultsContainer interface{}) (*AnalysisResponseContainer, error) {
	if a.Query.Tool == "" {
		return nil, fmt.Errorf("query contained empty tool, tool is required for rendering results: %+v", a.Query)
	}

	// Given a tool, we know what the container type _should_ be.
	// Take a moment to reflect and safe some panic.
	requiredContainer, err := c.vulnContainerForTool(a.Query.Tool)
	if err != nil {
		return nil, fmt.Errorf("tool '%s' unknown to api, cannot render", a.Query.Tool)
	}
	if reflect.PtrTo(reflect.TypeOf(requiredContainer)) != reflect.TypeOf(resultsContainer) {
		return nil, fmt.Errorf("expected output object type '%T', got '%T', cannot render", requiredContainer, resultsContainer)
	}

	resp := &AnalysisResponseContainer{}

	_, err = c.postResource(analysisEndpoint, a, resp)
	if err != nil {
		return nil, fmt.Errorf("analysis post failed: %w", err)
	}

	err = json.Unmarshal(resp.Results, resultsContainer)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Client) vulnContainerForTool(tool string) (interface{}, error) {
	// Return an empty object of the appropriate type for comparisons or initialization.
	switch tool {
	case "sumip":
		return []VulnSumIPResult{}, nil
	case "sumdnsname":
		return []VulnSumDNSNameResult{}, nil
	case "vulnipsummary":
		return []VulnIPSummaryResult{}, nil
	case "vulndetails":
		return []VulnDetailsResult{}, nil
	default:
		return nil, errors.New("can't identify an appropriate object for tool %s")
	}
}
