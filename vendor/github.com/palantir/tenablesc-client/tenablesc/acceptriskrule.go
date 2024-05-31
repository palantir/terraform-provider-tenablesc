package tenablesc

import (
	"encoding/json"
	"errors"
	"fmt"
)

const acceptRiskRuleEndpoint = "/acceptRiskRule"

// AcceptRiskRuleBaseFields are the fields renderable directly both to and from the API.
//
//	Requests take a list of repositories while all responses only ever return a single repository.
//	HostValue response structure also varies depending on type; the client masks out this conversion
//	 to and from the API structures.
type AcceptRiskRuleBaseFields struct {
	ID     string    `json:"id,omitempty"`
	Plugin *BaseInfo `json:"plugin,omitempty"`
	// HostType may be 'all', 'asset', 'ip', or 'uuid'
	HostType     string              `json:"hostType,omitempty"`
	Port         string              `json:"port,omitempty"`
	Protocol     string              `json:"protocol,omitempty"`
	Comments     string              `json:"comments,omitempty"`
	Expires      string              `json:"expires,omitempty"`
	Status       string              `json:"status,omitempty"`
	CreatedTime  UnixEpochStringTime `json:"createdTime,omitempty"`
	ModifiedTime UnixEpochStringTime `json:"modifiedTime,omitempty"`
}

type acceptRiskRuleInternal struct {
	AcceptRiskRuleBaseFields
	// ARR takes a list of repositories for writing rules.
	// It however only returns a single repository.
	Repository   *BaseInfo       `json:"repository,omitempty"`
	Repositories []BaseInfo      `json:"repositories,omitempty"`
	HostValue    json.RawMessage `json:"hostValue,omitempty"`
}

// AcceptRiskRule represents the Risk Rule structure in https://docs.tenable.com/tenablesc/api/Accept-Risk-Rule.htm
type AcceptRiskRule struct {
	AcceptRiskRuleBaseFields
	Repository *BaseInfo
	// HostValue will be rendered to/from the API-native structure as necessary based on HostType.
	HostValue string
}

func (a AcceptRiskRule) toInternal() (*acceptRiskRuleInternal, error) {
	rule := &acceptRiskRuleInternal{
		AcceptRiskRuleBaseFields: a.AcceptRiskRuleBaseFields,
		Repositories:             []BaseInfo{*a.Repository},
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

	return rule, nil
}

func (a acceptRiskRuleInternal) toExternal() (*AcceptRiskRule, error) {

	rule := &AcceptRiskRule{
		AcceptRiskRuleBaseFields: a.AcceptRiskRuleBaseFields,
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

	return rule, nil
}

func acceptRiskRuleSliceToExternal(a []acceptRiskRuleInternal) ([]*AcceptRiskRule, error) {
	rules := make([]*AcceptRiskRule, 0, len(a))

	for _, a := range a {
		rule, err := a.toExternal()
		if err != nil {
			return nil, fmt.Errorf("failed to convert to external struct: %w", err)
		}

		rules = append(rules, rule)
	}
	return rules, nil
}

func (c *Client) GetAllAcceptRiskRules() ([]*AcceptRiskRule, error) {
	var a []acceptRiskRuleInternal

	if _, err := c.getResource(acceptRiskRuleEndpoint, &a); err != nil {
		return nil, fmt.Errorf("failed to get accept risk rules: %w", err)
	}
	return acceptRiskRuleSliceToExternal(a)
}

func (c *Client) CreateAcceptRiskRule(a *AcceptRiskRule) (*AcceptRiskRule, error) {
	var resp []*acceptRiskRuleInternal

	aInt, err := a.toInternal()
	if err != nil {
		return nil, fmt.Errorf("failed to parse rule to internal format: %w", err)
	}

	if _, err := c.postResource(acceptRiskRuleEndpoint, aInt, &resp); err != nil {
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

func (c *Client) GetAcceptRiskRule(id string) (*AcceptRiskRule, error) {
	resp := &acceptRiskRuleInternal{}

	if _, err := c.getResource(fmt.Sprintf("%s/%s", acceptRiskRuleEndpoint, id), resp); err != nil {
		return nil, fmt.Errorf("failed to get accept risk rule id %s: %w", id, err)
	}

	return resp.toExternal()
}

func (c *Client) DeleteAcceptRiskRule(id string) error {

	if _, err := c.deleteResource(fmt.Sprintf("%s/%s", acceptRiskRuleEndpoint, id), nil, nil); err != nil {
		return fmt.Errorf("unable to delete accept risk rule with id %s: %w", id, err)
	}

	return nil
}
