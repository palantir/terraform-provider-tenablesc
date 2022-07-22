// Copyright 2022 Palantir Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/palantir/tenablesc-client/tenablesc"
)

func ResourceRepositoryOrganizationAssociation() *schema.Resource {
	return &schema.Resource{
		Description:   descriptionResourceRepositoryOrganizationAssociation,
		CreateContext: resourceRepositoryOrganizationAssociationCreateOrUpdate,
		ReadContext:   resourceRepositoryOrganizationAssociationRead,
		UpdateContext: resourceRepositoryOrganizationAssociationCreateOrUpdate,
		DeleteContext: resourceRepositoryOrganizationAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"repository_id": {
				Type:        schema.TypeString,
				Description: descriptionRepositoryID,
				Required:    true,
			},
			"organization": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"organization_id": {
							Type:        schema.TypeInt,
							Description: descriptionOrganizationID,
							Required:    true,
						},
						"group_assignment": {
							Type:             schema.TypeString,
							Description:      descriptionGroupAssignment,
							Optional:         true,
							ValidateDiagFunc: validateGroupAssignment,
						},
					},
				},
				Required: true,
			},
		},
	}
}

func resourceRepositoryOrganizationAssociationCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")

	sc := m.(*tenablesc.Client)

	d.SetId(d.Get("repository_id").(string))

	repositoryAssociationInput, err := buildRepositoryOrganizationAssociationInputs(d)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = sc.UpdateRepository(repositoryAssociationInput)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRepositoryOrganizationAssociationRead(ctx, d, m)
}

func resourceRepositoryOrganizationAssociationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	repository, err := sc.GetRepository(d.Id())
	if err != nil {
		return handleNotFoundError(d, err)
	}

	Logf(logDebug, "response: %+v", repository)

	d.SetId(string(repository.ID))

	organizations := make([]map[string]interface{}, 0)
	for _, ro := range repository.Organizations {
		org := make(map[string]interface{})

		orgIDInt, err := strconv.ParseInt(ro.ID, 10, 32)
		if err != nil {
			return diag.FromErr(err)
		}
		org["org_id"] = int(orgIDInt)
		org["group_assignment"] = ro.GroupAssign

		organizations = append(organizations, org)
	}
	d.Set("organization", organizations)

	return nil
}

func resourceRepositoryOrganizationAssociationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sc := m.(*tenablesc.Client)

	repositoryAssociation := &tenablesc.Repository{
		RepoBaseFields: tenablesc.RepoBaseFields{
			BaseInfo: tenablesc.BaseInfo{
				ID: tenablesc.ProbablyString(d.Id()),
			},
			Organizations: []tenablesc.RepoOrganization{},
		},
	}

	_, err := sc.UpdateRepository(repositoryAssociation)
	if err != nil {
		return handleNotFoundError(d, err)
	}
	return nil
}

func buildRepositoryOrganizationAssociationInputs(d *schema.ResourceData) (*tenablesc.Repository, error) {

	repoID := d.Id()
	organizations := d.Get("organization").([]interface{})

	repositoryAssociation := &tenablesc.Repository{
		RepoBaseFields: tenablesc.RepoBaseFields{
			BaseInfo: tenablesc.BaseInfo{
				ID: tenablesc.ProbablyString(repoID),
			},
		},
	}

	var repoOrgs []tenablesc.RepoOrganization

	for _, org := range organizations {
		org := org.(map[string]interface{})

		id, ok := org["org_id"]
		if !ok {
			return nil, fmt.Errorf("got null organization id")
		}

		groupAssignment, ok := org["group_assignment"]
		if !ok {
			groupAssignment = ""
		}

		repoOrgs = append(repoOrgs, tenablesc.RepoOrganization{
			ID:          strconv.Itoa(id.(int)),
			GroupAssign: groupAssignment.(string),
		})
	}

	repositoryAssociation.Organizations = repoOrgs

	repositoryAssociationBytes, err := json.Marshal(repositoryAssociation)
	if err != nil {
		return nil, err

	}
	Logf(logDebug, fmt.Sprintf("built Repository Association: %s", repositoryAssociationBytes))

	return repositoryAssociation, nil
}

var validGroupAssignments = []string{
	"",
	"all",
	"fullAccess",
	"partial",
}

func validateGroupAssignment(i interface{}, path cty.Path) diag.Diagnostics {

	if i, ok := i.(string); ok {
		for _, v := range validGroupAssignments {
			if i == v {
				return nil
			}
		}
	}

	return diag.Errorf("%v is not a valid value for group_assignment", i)

}
