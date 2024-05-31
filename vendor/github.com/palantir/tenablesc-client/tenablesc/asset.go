package tenablesc

import (
	"fmt"
	"strings"
)

const assetsEndpoint = "/asset"

// Asset defines common fields used in requests and responses for assets
//
//	per https://docs.tenable.com/tenablesc/api/Asset.htm
//
// Some fields are presented abstracted here and internally reformatted for transport.
type Asset struct {
	BaseInfo
	Type    string   `json:"type,omitempty"`
	Prepare FakeBool `json:"prepare,omitempty"`
	// DefinedDNSNames is a comma-separated list on the wire; this is translated during marshal/unmarshal.
	DefinedDNSNames []string `json:"-"`
	// DefinedIPs is a comma-separated list on the wire; this is translated during marshal/unmarshal.
	DefinedIPs   []string            `json:"-"`
	IPCount      ProbablyString      `json:"ipCount,omitempty"`
	CreatedTime  UnixEpochStringTime `json:"createdTime,omitempty"`
	ModifiedTime UnixEpochStringTime `json:"modifiedTime,omitempty"`
	// Repositories includes a redundant `IPCount` field on the wire; this is dropped during marshal/unmarshal for consistency.
	Repositories []Repository `json:"-"`
}

type assetRequest struct {
	Asset
	DefinedDNSNames string `json:"definedDNSNames,omitempty"`
	DefinedIPs      string `json:"definedIPs,omitempty"`
}

func assetFromExternal(a *Asset) *assetRequest {
	return &assetRequest{
		Asset:           *a,
		DefinedDNSNames: strings.Join(a.DefinedDNSNames, ","),
		DefinedIPs:      strings.Join(a.DefinedIPs, ","),
	}
}

type assetResponse struct {
	Asset
	TypeFields struct {
		DefinedDNSNames string `json:"definedDNSNames,omitempty"`
		DefinedIPs      string `json:"definedIPs,omitempty"`
	} `json:"typeFields,omitempty"`
	Repositories []struct {
		IPCount    string     `json:"ipCount,omitempty"`
		Repository Repository `json:"repository,omitempty"`
	} `json:"repositories,omitempty"`
}

type orgAssetResponse struct {
	Manageable []*assetResponse `json:"manageable,omitempty" tenable:"recurse"`
	Usable     []*assetResponse `json:"usable,omitempty" tenable:"recurse"`
}

func (a assetResponse) toExternal() *Asset {
	as := &a.Asset

	as.DefinedDNSNames = strings.Split(a.TypeFields.DefinedDNSNames, ",")
	as.DefinedIPs = strings.Split(a.TypeFields.DefinedIPs, ",")

	for _, r := range a.Repositories {
		as.Repositories = append(as.Repositories, r.Repository)
	}

	return as
}

func assetSliceToExternal(ai []*assetResponse) ([]*Asset, error) {
	assets := make([]*Asset, 0, len(ai))

	for _, a := range ai {
		asset := a.toExternal()
		assets = append(assets, asset)

	}
	return assets, nil
}

func orgAssetSliceToExternal(o orgAssetResponse) []*Asset {
	var orgOut []*Asset
	orgMap := make(map[ProbablyString]bool)

	for _, o := range o.Usable {
		orgMap[o.ID] = true
		orgOut = append(orgOut, o.toExternal())
	}
	for _, o := range o.Manageable {
		if _, exists := orgMap[o.ID]; !exists {
			orgOut = append(orgOut, o.toExternal())
			orgMap[o.ID] = true
		}
	}

	return orgOut
}

func (c *Client) GetAllAssets() ([]*Asset, error) {
	// We get a different struct depending on whether we're
	// an org or admin user, but only for get all, not individual assets.
	var orgResponse orgAssetResponse
	_, err := c.getResource(assetsEndpoint, &orgResponse)
	if err == nil {
		return orgAssetSliceToExternal(orgResponse), nil
	}

	// okay, didn't go smooth, assume we're an admin user.
	// Using this ordering because we're far more likely to be managing assets at the org level.
	var resp []*assetResponse
	_, err = c.getResource(assetsEndpoint, &resp)
	if err != nil {
		return nil, err
	}

	return assetSliceToExternal(resp)
}

func (c *Client) CreateAsset(a *Asset) (*Asset, error) {
	resp := &assetResponse{}

	if _, err := c.postResource(assetsEndpoint, assetFromExternal(a), resp); err != nil {
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	return resp.toExternal(), nil
}

func (c *Client) GetAsset(id string) (*Asset, error) {
	resp := &assetResponse{}

	if _, err := c.getResource(fmt.Sprintf("%s/%s", assetsEndpoint, id), resp); err != nil {
		return nil, fmt.Errorf("failed to get asset id %s: %w", id, err)
	}

	return resp.toExternal(), nil
}

func (c *Client) UpdateAsset(a *Asset) (*Asset, error) {
	resp := &assetResponse{}

	if _, err := c.patchResourceWithID(assetsEndpoint, assetFromExternal(a), resp); err != nil {
		return nil, fmt.Errorf("failed to update asset: %w", err)
	}

	return resp.toExternal(), nil
}

func (c *Client) DeleteAsset(id string) error {
	if _, err := c.deleteResource(fmt.Sprintf("%s/%s", assetsEndpoint, id), nil, nil); err != nil {
		return fmt.Errorf("failed to delete asset %s: %w", id, err)
	}
	return nil
}
