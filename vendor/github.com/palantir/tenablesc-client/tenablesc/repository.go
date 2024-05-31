package tenablesc

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const reposEndpoint = "/repository"

// Repository represents the fields for https://docs.tenable.com/tenablesc/api/Repository.htm
// Each repository type has a significantly different structure that is rendered as needed.
type Repository struct {
	RepoBaseFields
	RepoFieldsCommon
	RepoIPFields
}

// RepoFieldsCommon includes the fields common to requests and responses in this endpoint for all repository types.
type RepoFieldsCommon struct {
	ActiveVulnsLifetime     string   `json:"activeVulnsLifetime,omitempty"`
	ComplianceVulnsLifetime string   `json:"complianceVulnsLifetime,omitempty"`
	MitigatedVulnsLifetime  string   `json:"mitigatedVulnsLifetime,omitempty"`
	TrendingDays            string   `json:"trendingDays,omitempty"`
	TrendWithRaw            FakeBool `json:"trendWithRaw,omitempty"`
}

// RepoIPFields includes the fields only available in IPv4/6 repositories and not Agent-based repositories.
type RepoIPFields struct {
	IPRange                string              `json:"ipRange,omitempty"`
	IPCount                string              `json:"ipCount,omitempty"`
	PassiveVulnsLifetime   string              `json:"passiveVulnsLifetime,omitempty"`
	LCEVulnsLifetime       string              `json:"lveVulnsLifetime,omitempty"`
	LastGenerateNessusTime UnixEpochStringTime `json:"lastGenerateNessusTime,omitempty"`
	NessusSchedule         *NessusSchedule     `json:"nessusSchedule,omitempty"`
}

// RepoBaseFields includes the Repository fields common to responses from this endpoint and others.
type RepoBaseFields struct {
	BaseInfo
	DataFormat     string              `json:"dataFormat,omitempty"`
	Type           string              `json:"type,omitempty"`
	DownloadFormat string              `json:"downloadFormat,omitempty"`
	CreatedTime    UnixEpochStringTime `json:"createdTime,omitempty"`
	ModifiedTime   UnixEpochStringTime `json:"modifiedTime,omitempty"`
	Running        FakeBool            `json:"running,omitempty"`
	Organizations  []RepoOrganization  `json:"organizations,omitempty"`
}

type NessusSchedule struct {
	Type       string `json:"type,omitempty"`
	Start      string `json:"start,omitempty"`
	RepeatRule string `json:"repeatRule,omitempty"`
}

type RepoOrganization struct {
	ID          string `json:"id,omitempty"`
	GroupAssign string `json:"groupAssign,omitempty"`
}

// input and output formats are different.  Handle the differences internally
type repoInternal struct {
	RepoBaseFields
	TypeFields json.RawMessage `json:"typeFields,omitempty"`
}

func (r repoInternal) toExternal() (*Repository, error) {
	if r.Type != "Local" {
		return nil, fmt.Errorf("Repo type %s is not supported.  Only local is", r.Type)
	}

	repo := &Repository{
		RepoBaseFields: r.RepoBaseFields,
	}

	if r.DataFormat == "mobile" {
		return nil, errors.New("mobile repo data format is not supported")
	}

	if err := json.Unmarshal(r.TypeFields, &repo.RepoFieldsCommon); err != nil {
		return nil, fmt.Errorf("faild to unmarshal typeFields: %w", err)
	}

	if strings.HasPrefix(r.DataFormat, "IP") {
		if err := json.Unmarshal(r.TypeFields, &repo.RepoIPFields); err != nil {
			return nil, fmt.Errorf("faild to unmarshal typeFields: %w", err)
		}
	}

	return repo, nil
}

func repoSliceToExternal(r []repoInternal) ([]*Repository, error) {
	repos := make([]*Repository, 0, len(r))

	for _, r := range r {
		repo, err := r.toExternal()
		if err != nil {
			return nil, fmt.Errorf("failed to convert to external struct: %w", err)
		}

		repos = append(repos, repo)
	}
	return repos, nil
}

func (c *Client) GetAllRepositories() ([]*Repository, error) {
	var r []repoInternal

	if _, err := c.getResource(reposEndpoint, &r); err != nil {
		return nil, fmt.Errorf("failed to get repositories: %w", err)
	}
	return repoSliceToExternal(r)
}

func (c *Client) CreateRepository(r *Repository) (*Repository, error) {
	resp := &repoInternal{}

	if _, err := c.postResource(reposEndpoint, r, resp); err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	return resp.toExternal()

}

func (c *Client) GetRepository(id string) (*Repository, error) {
	resp := &repoInternal{}

	if _, err := c.getResource(fmt.Sprintf("%s/%s", reposEndpoint, id), resp); err != nil {
		return nil, fmt.Errorf("failed to get repo id %s: %w", id, err)
	}

	return resp.toExternal()
}

func (c *Client) UpdateRepository(repo *Repository) (*Repository, error) {
	resp := &repoInternal{}

	if _, err := c.patchResourceWithID(reposEndpoint, repo, resp); err != nil {
		return nil, fmt.Errorf("failed to update repo: %w", err)
	}

	return resp.toExternal()
}

func (c *Client) DeleteRepository(id string) error {
	if _, err := c.deleteResource(fmt.Sprintf("%s/%s", reposEndpoint, id), nil, nil); err != nil {
		return fmt.Errorf("failed to delete repo %s, %w", id, err)
	}
	return nil
}
