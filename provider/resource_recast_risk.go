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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/palantir/tenablesc-client/tenablesc"
)

// ResourceRecastRisk Initialize the Recast Resource
func ResourceRecastRisk() *schema.Resource {
	return &schema.Resource{
		Description:   descriptionResourceRecastRisk,
		CreateContext: resourceRecastRiskCreate,
		ReadContext:   resourceRecastRiskRead,
		UpdateContext: resourceRecastRiskUpdate,
		DeleteContext: resourceRecastRiskDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"repository_id": {
				Type:        schema.TypeString,
				Description: descriptionRepositoryID,
				Required:    true,
			},
			"plugin_id": {
				Type:        schema.TypeString,
				Description: descriptionPluginID,
				Required:    true,
				ForceNew:    true,
			},
			"new_severity": {
				Type:        schema.TypeString,
				Description: descriptionRecastNewSeverity,
				Optional:    true,
				Default:     "0",
			},
			"host_type": {
				Type:        schema.TypeString,
				Description: descriptionRiskRuleHostType,
				Optional:    true,
				Default:     "all",
			},
			"host_value": {
				Type:                  schema.TypeString,
				Description:           descriptionRiskRuleHostValue,
				Optional:              true,
				Default:               "",
				DiffSuppressOnRefresh: true,
				DiffSuppressFunc:      diffSuppressNormalizedIPSet,
			},
			"port": {
				Type:        schema.TypeString,
				Description: descriptionPort,
				Optional:    true,
				Default:     "any",
			},
			"protocol": {
				Type:        schema.TypeString,
				Description: descriptionProtocol,
				Optional:    true,
				Default:     "any",
			},
			"comments": {
				Type:        schema.TypeString,
				Description: descriptionComments,
				Optional:    true,
				Default:     descriptionDefaultDescriptionValue,
			},
		},
	}
}

func resourceRecastRiskCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")

	sc := m.(*tenablesc.Client)

	recastRiskResponse, err := sc.CreateRecastRiskRule(buildRecastRiskInput(d))
	if err != nil {
		return diag.FromErr(err)
	}

	Logf(logDebug, "response: %+v", recastRiskResponse)

	d.SetId(recastRiskResponse.ID)

	return resourceRecastRiskRead(ctx, d, m)
}

func resourceRecastRiskRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")

	sc := m.(*tenablesc.Client)

	recastRiskResponse, err := sc.GetRecastRiskRule(d.Id())
	if err != nil {
		return handleNotFoundError(d, err)
	}

	d.Set("plugin_id", recastRiskResponse.Plugin.ID)

	Logf(logDebug, "response: %+v", recastRiskResponse)

	hostType := recastRiskResponse.HostType
	d.Set("host_type", hostType)
	d.Set("host_value", recastRiskResponse.HostValue)
	d.Set("new_severity", recastRiskResponse.NewSeverity)
	d.Set("host_type", recastRiskResponse.HostType)
	d.Set("port", recastRiskResponse.Port)
	d.Set("protocol", recastRiskResponse.Protocol)
	d.Set("comments", recastRiskResponse.Comments)
	d.Set("repository_id", recastRiskResponse.Repository.ID)
	d.SetId(recastRiskResponse.ID)

	return nil
}

func resourceRecastRiskUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")

	deleteError := resourceRecastRiskDelete(ctx, d, m)
	if deleteError != nil {
		return deleteError
	}

	createError := resourceRecastRiskCreate(ctx, d, m)
	if createError != nil {
		return createError
	}

	return resourceRecastRiskRead(ctx, d, m)
}

func resourceRecastRiskDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")

	sc := m.(*tenablesc.Client)

	err := sc.DeleteRecastRiskRule(d.Id())
	if err != nil {
		return handleNotFoundError(d, err)
	}
	return nil
}

func buildRecastRiskInput(d *schema.ResourceData) *tenablesc.RecastRiskRule {
	pluginID := d.Get("plugin_id").(string)
	repositoryID := d.Get("repository_id").(string)
	newSeverity := d.Get("new_severity").(string)
	hostType := d.Get("host_type").(string)
	hostValue := d.Get("host_value").(string)
	port := d.Get("port").(string)
	protocol := d.Get("protocol").(string)
	comments := d.Get("comments").(string)

	rrInput := &tenablesc.RecastRiskRule{
		RecastRiskRuleBaseFields: tenablesc.RecastRiskRuleBaseFields{
			ID:       d.Id(),
			Plugin:   tenablesc.BaseInfo{ID: tenablesc.ProbablyString(pluginID)},
			Port:     port,
			Protocol: protocol,
			Comments: comments,
			HostType: hostType,
		},
		Repository:  tenablesc.BaseInfo{ID: tenablesc.ProbablyString(repositoryID)},
		NewSeverity: newSeverity,
		HostValue:   hostValue,
	}

	return rrInput
}
