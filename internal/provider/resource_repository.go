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
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/palantir/tenablesc-client/tenablesc"
)

func ResourceRepository() *schema.Resource {
	return &schema.Resource{
		Description:   descriptionResourceRepository,
		CreateContext: resourceRepositoryCreate,
		ReadContext:   resourceRepositoryRead,
		UpdateContext: resourceRepositoryUpdate,
		DeleteContext: resourceRepositoryDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: descriptionRepositoryName,
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: descriptionRepositoryDescription,
				Optional:    true,
				Default:     descriptionDefaultDescriptionValue,
			},
			"ip_range": {
				Type:        schema.TypeString,
				Description: descriptionRepositoryIPRange,
				Required:    true,
			},
			"trending_days": {
				Type:        schema.TypeInt,
				Description: descriptionTrendingDays,
				Optional:    true,
				Default:     30,
			},
			"trend_with_raw": {
				Type:        schema.TypeBool,
				Description: descriptionTrendWithRaw,
				Optional:    true,
				Default:     false,
			},
			"vulnerability_lifetimes": {
				Type:        schema.TypeList,
				Description: descriptionVulnerabilityLifetime,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"active_days": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"passive_days": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"compliance_days": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"mitigated_days": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
		},
	}

}

func resourceRepositoryCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")

	sc := m.(*tenablesc.Client)

	repo, err := sc.CreateRepository(buildRepoInputs(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(string(repo.ID))

	return resourceRepositoryRead(ctx, d, m)
}

func resourceRepositoryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")

	sc := m.(*tenablesc.Client)

	repository, err := sc.GetRepository(d.Id())
	if err != nil {
		d.SetId("")
		return handleNotFoundError(d, err)
	}

	Logf(logDebug, "response: %+v", repository)

	d.SetId(string(repository.ID))
	d.Set("name", repository.Name)
	d.Set("description", repository.Description)
	d.Set("ip_range", repository.IPRange)
	d.Set("trending_days", repository.TrendingDays)
	d.Set("trend_with_raw", repository.TrendWithRaw.AsBool())

	vulnLifetimes := make(map[string]interface{})

	if repository.ComplianceVulnsLifetime != "" {
		v, err := strconv.ParseInt(repository.ComplianceVulnsLifetime, 10, 32)
		if err != nil {
			return diag.FromErr(err)
		}
		vulnLifetimes["compliance_days"] = int(v)
	}

	if repository.MitigatedVulnsLifetime != "" {
		v, err := strconv.ParseInt(repository.MitigatedVulnsLifetime, 10, 32)
		if err != nil {
			return diag.FromErr(err)
		}
		vulnLifetimes["mitigated_days"] = int(v)
	}

	if repository.ActiveVulnsLifetime != "" {
		v, err := strconv.ParseInt(repository.ActiveVulnsLifetime, 10, 32)
		if err != nil {
			return diag.FromErr(err)
		}
		vulnLifetimes["active_days"] = int(v)
	}

	if repository.ActiveVulnsLifetime != "" {
		v, err := strconv.ParseInt(repository.ActiveVulnsLifetime, 10, 32)
		if err != nil {
			return diag.FromErr(err)
		}
		vulnLifetimes["passive_days"] = int(v)
	}

	if len(vulnLifetimes) > 0 {
		d.Set("vulnerability_lifetimes", vulnLifetimes)
	}

	return nil
}

func resourceRepositoryDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")

	sc := m.(*tenablesc.Client)

	err := sc.DeleteRepository(d.Id())
	if err != nil {
		return handleNotFoundError(d, err)
	}

	return nil
}

func resourceRepositoryUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	_, err := sc.UpdateRepository(buildRepoInputs(d))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRepositoryRead(ctx, d, m)
}

func buildRepoInputs(d *schema.ResourceData) *tenablesc.Repository {

	name := d.Get("name").(string)
	description := d.Get("description").(string)

	repo := &tenablesc.Repository{
		RepoBaseFields: tenablesc.RepoBaseFields{
			BaseInfo: tenablesc.BaseInfo{
				ID:          tenablesc.ProbablyString(d.Id()),
				Name:        name,
				Description: description,
			},
			DataFormat: "IPv4",  // will need to be updated when we add agent repos.
			Type:       "Local", // Other options: remote, offline.
		},
		RepoFieldsCommon: tenablesc.RepoFieldsCommon{},
		RepoIPFields:     tenablesc.RepoIPFields{},
	}

	repo.IPRange = d.Get("ip_range").(string)
	repo.TrendingDays = strconv.Itoa(d.Get("trending_days").(int))
	repo.TrendWithRaw = tenablesc.ToFakeBool(d.Get("trend_with_raw").(bool))

	if vulnLifetimes, ok := d.GetOk("vulnerability_lifetimes"); ok {
		vulnLifetimes := vulnLifetimes.([]interface{})
		if len(vulnLifetimes) > 0 {
			vulnLifetimes := vulnLifetimes[0].(map[string]interface{})

			if v, ok := vulnLifetimes["active_days"]; ok {
				repo.ActiveVulnsLifetime = strconv.Itoa(v.(int))
			}
			if v, ok := vulnLifetimes["passive_days"]; ok {
				repo.PassiveVulnsLifetime = strconv.Itoa(v.(int))
			}
			if v, ok := vulnLifetimes["mitigated_days"]; ok {
				repo.MitigatedVulnsLifetime = strconv.Itoa(v.(int))
			}
			if v, ok := vulnLifetimes["compliance_days"]; ok {
				repo.ComplianceVulnsLifetime = strconv.Itoa(v.(int))
			}
		}
	}

	return repo
}
