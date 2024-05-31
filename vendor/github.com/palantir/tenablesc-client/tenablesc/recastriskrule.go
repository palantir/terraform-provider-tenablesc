package tenablesc

import (
	"encoding/json"
	"errors"
	"fmt"
)

const recastRiskRuleEndpoint = "/recastRiskRule"

const (
	RecastSeverityInfo   = "0"
	RecastSeverityLow    = "1"
	RecastSeverityMedium = "2"
	RecastSeverityHigh   = "3"
)

// RecastRiskRuleBaseFields are the fields renderable directly both to and from the API.
//
//	Requests take a list of repositories while all responses only ever return a single repository.
//	HostValue response structure also varies depending on type; the client masks out this conversion
//	 to and from the API structures.
type RecastRiskRuleBaseFields struct {
	ID           string   `json:"id,omitempty"`
	Organization BaseInfo `json:"organization,omitempty"`
	User         UserInfo `json:"user,omitempty"`
	Plugin       BaseInfo `json:"plugin,omitempty"`
	// HostType may be 'all', 'asset', 'ip', or 'uuid'
	HostType     string              `json:"hostType,omitempty"`
	Port         string              `json:"port,omitempty"`
	Protocol     string              `json:"protocol,omitempty"`
	Order        string              `json:"order,omitempty"`
	Comments     string              `json:"comments,omitempty"`
	Expires      string              `json:"expires,omitempty"`
	Status       string              `json:"status,omitempty"`
	CreatedTime  UnixEpochStringTime `json:"createdTime,omitempty"`
	ModifiedTime UnixEpochStringTime `json:"modifiedTime,omitempty"`
}

type recastRiskRuleInternal struct {
	RecastRiskRuleBaseFields
	// RRR takes a list of repositories for writing rules.
	// It however only returns a single repository.
	NewSeverity  json.RawMessage `json:"newSeverity,omitempty"`
	Repository   BaseInfo        `json:"repository,omitempty"`
	Repositories []BaseInfo      `json:"repositories,omitempty"`
	HostValue    json.RawMessage `json:"hostValue,omitempty"`
}

// RecastRiskRule represents the Risk Rule structure in https://docs.tenable.com/tenablesc/api/Recast-Risk-Rule.htm
type RecastRiskRule struct {
	RecastRiskRuleBaseFields
	NewSeverity string `json:"newSeverity,omitempty"`
	Repository  BaseInfo
	HostValue   string
}

func (a RecastRiskRule) toInternal() (*recastRiskRuleInternal, error) {
	rule := &recastRiskRuleInternal{
		RecastRiskRuleBaseFields: a.RecastRiskRuleBaseFields,
		Repositories:             []BaseInfo{a.Repository},
	}
	switch a.HostType {
	case "all":
		rule.HostValue = []byte("\"\"")
	case "asset":
		assetInfo := BaseInfo{ID: ProbablyString(a.HostValue)}
		assetBytes, err := json.Marshal(&assetInfo)
		if err != nil {
			return nil, fmt.Errorf("could not parse asset hostvalue %s into id struct", a.HostValue)
		}
		rule.HostValue = assetBytes
	case "ip", "uuid":
		// Convert this to a legal json string; this does nice things like escape newlines properly.
		assetBytes, err := json.Marshal(&a.HostValue)
		if err != nil {
			return nil, fmt.Errorf("could not parse asset hostvalue %s into json string", a.HostValue)
		}
		rule.HostValue = assetBytes
	default:
		return nil, fmt.Errorf("HostType %s not supported in client", a.HostType)
	}

	newSeverityBytes, err := json.Marshal(&BaseInfo{ID: ProbablyString(a.NewSeverity)})
	if err != nil {
		return nil, fmt.Errorf("could not parse newSeverity as string")
	}
	rule.NewSeverity = newSeverityBytes

	return rule, nil
}

func (a recastRiskRuleInternal) toExternal() (*RecastRiskRule, error) {

	rule := &RecastRiskRule{
		RecastRiskRuleBaseFields: a.RecastRiskRuleBaseFields,
		Repository:               a.Repository,
	}

	switch a.HostType {
	case "all":
		rule.HostValue = ""
	case "asset":
		assetInfo := BaseInfo{}
		if err := json.Unmarshal(a.HostValue, &assetInfo); err != nil {
			return nil, fmt.Errorf("could not parse asset hostvalue %s as id struct", string(a.HostValue))
		}
		rule.HostValue = string(assetInfo.ID)
	case "ip", "uuid":
		rule.HostValue = string(a.HostValue)
	default:
		return nil, fmt.Errorf("HostType %s not supported in client", a.HostType)
	}

	var newSeverityString string
	if err := json.Unmarshal(a.NewSeverity, &newSeverityString); err != nil {
		return nil, fmt.Errorf("unable to unmarshal NewSeverity '%s' as string", a.NewSeverity)
	}

	rule.NewSeverity = newSeverityString

	return rule, nil
}

func recastRiskRuleSliceToExternal(r []recastRiskRuleInternal) ([]*RecastRiskRule, error) {
	rules := make([]*RecastRiskRule, 0, len(r))

	for _, r := range r {
		rule, err := r.toExternal()
		if err != nil {
			return nil, fmt.Errorf("failed to convert to external struct: %w", err)
		}

		rules = append(rules, rule)
	}
	return rules, nil
}

func (c *Client) GetAllRecastRiskRules() ([]*RecastRiskRule, error) {
	var a []recastRiskRuleInternal

	if _, err := c.getResource(recastRiskRuleEndpoint, &a); err != nil {
		return nil, fmt.Errorf("failed to get recast risk rules: %w", err)
	}
	return recastRiskRuleSliceToExternal(a)
}

func (c *Client) CreateRecastRiskRule(a *RecastRiskRule) (*RecastRiskRule, error) {
	var resp []*recastRiskRuleInternal

	aInt, err := a.toInternal()
	if err != nil {
		return nil, fmt.Errorf("failed to parse rule to internal format: %w", err)
	}

	if _, err := c.postResource(recastRiskRuleEndpoint, aInt, &resp); err != nil {
		return nil, fmt.Errorf("failed to create recast risk rule: %w", err)
	}

	if len(resp) == 0 {
		return nil, errors.New("got zero length response to create recast")
	}
	if len(resp) > 1 {
		return nil, errors.New("got multiple responses to create recast")
	}

	return resp[0].toExternal()
}

func (c *Client) GetRecastRiskRule(id string) (*RecastRiskRule, error) {
	resp := &recastRiskRuleInternal{}

	if _, err := c.getResource(fmt.Sprintf("%s/%s", recastRiskRuleEndpoint, id), resp); err != nil {
		return nil, fmt.Errorf("failed to get recast risk rule id %s: %w", id, err)
	}

	return resp.toExternal()
}

func (c *Client) DeleteRecastRiskRule(id string) error {

	if _, err := c.deleteResource(fmt.Sprintf("%s/%s", recastRiskRuleEndpoint, id), nil, nil); err != nil {
		return fmt.Errorf("unable to delete recast risk rule with id %s: %w", id, err)
	}

	return nil
}
