package tenablesc

import (
	"fmt"
	"net/url"
)

const pluginEndpoint = "/plugin"

// Plugin represents the response structure for https://docs.tenable.com/tenablesc/api/Plugin.htm
type Plugin struct {
	BaseInfo
	Family              Family         `json:"family,omitempty"`
	CVSSV3BaseScore     string         `json:"cvssV3BaseScore,omitempty"`
	CVSSV3TemporalScore string         `json:"cvssV3TemporalScore,omitempty"`
	PluginPubDate       ProbablyString `json:"pluginPubDate,omitempty"`
	PluginModDate       ProbablyString `json:"pluginModDate,omitempty"`
	PatchPubDate        ProbablyString `json:"patchPubDate,omitempty"`
	PatchModDate        ProbablyString `json:"patchModDate,omitempty"`
	VulnPubDate         ProbablyString `json:"vulnPubDate,omitempty"`
	ModifiedTime        ProbablyString `json:"modifiedTime,omitempty"`
	CheckType           string         `json:"checkType,omitempty"`
}

type Family struct {
	BaseInfo
	Type string `json:"type,omitempty"`
}

func (c *Client) GetPlugin(id string) (*Plugin, error) {
	plugin := &Plugin{}

	if _, err := c.getResource(fmt.Sprintf("%s/%s", pluginEndpoint, id), plugin); err != nil {
		return nil, fmt.Errorf("failed to get plugin id %s: %w", id, err)
	}

	return plugin, nil
}

func (c *Client) GetPluginsByName(name string) ([]*Plugin, error) {
	var plugins []*Plugin

	query := url.Values{}
	query.Add("sortDirection", "ASC")
	query.Add("sortField", "name")
	query.Add("filterField", "name")
	query.Add("op", "eq")
	query.Add("value", name)

	if _, err := c.getResource(fmt.Sprintf("%s?%s", pluginEndpoint, query.Encode()), &plugins); err != nil {
		return nil, fmt.Errorf("failed to find plugin with name %s: %w", name, err)
	}

	return plugins, nil
}
