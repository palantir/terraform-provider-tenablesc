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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/palantir/tenablesc-client/tenablesc"
)

const timeLayout = "2006-01-02T15:04:00Z07:00"

// ResourceAcceptRisk Initialize the Accept Risk Resource
func ResourceAcceptRisk() *schema.Resource {
	return &schema.Resource{
		Description:   descriptionResourceAcceptRisk,
		CreateContext: resourceAcceptRiskCreate,
		ReadContext:   resourceAcceptRiskRead,
		UpdateContext: resourceAcceptRiskUpdate,
		DeleteContext: resourceAcceptRiskDelete,

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
				DiffSuppressFunc:      diffSuppressNormalizedIPSet,
				DiffSuppressOnRefresh: true,
			},
			"port": {
				Type:        schema.TypeString,
				Description: descriptionPort,
				Optional:    true,
				Default:     "any",
			},
			"protocol": {
				Type:             schema.TypeString,
				Description:      descriptionProtocol,
				Optional:         true,
				Default:          "any",
				ValidateDiagFunc: validateRecastAcceptRiskProtocol,
			},
			"expiration": {
				Type:                  schema.TypeString,
				Description:           descriptionAcceptRiskExpiration,
				Optional:              true,
				DiffSuppressFunc:      DiffSuppressParsedTimes,
				DiffSuppressOnRefresh: true,
				Default:               "-1",
				ValidateDiagFunc:      validateAcceptRiskExpirationInFuture,
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

func resourceAcceptRiskCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")

	sc := m.(*tenablesc.Client)

	repositoryID := d.Get("repository_id").(string)
	pluginID := d.Get("plugin_id").(string)
	hostType := d.Get("host_type").(string)
	hostValue := d.Get("host_value").(string)
	port := d.Get("port").(string)
	expiration := d.Get("expiration").(string)
	comments := d.Get("comments").(string)

	protocol, err := getRecastAcceptRiskProtocolID(d.Get("protocol").(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get protocol id: %w", err))
	}

	rule := &tenablesc.AcceptRiskRule{
		AcceptRiskRuleBaseFields: tenablesc.AcceptRiskRuleBaseFields{
			Plugin:   &tenablesc.BaseInfo{ID: tenablesc.ProbablyString(pluginID)},
			HostType: hostType,
			Port:     port,
			Protocol: protocol,
			Comments: comments,
		},
		Repository: &tenablesc.BaseInfo{ID: tenablesc.ProbablyString(repositoryID)},
		HostValue:  hostValue,
	}
	if strings.Compare(expiration, "-1") != 0 {
		expirationTime, parseErr := time.Parse(timeLayout, expiration)
		if parseErr != nil {
			return diag.FromErr(parseErr)
		}

		et := fmt.Sprintf("%d", expirationTime.Unix())
		rule.Expires = et
	}

	acceptRisks, err := sc.CreateAcceptRiskRule(rule)
	if err != nil {
		return diag.FromErr(err)
	}

	Logf(logDebug, "response: %+v", acceptRisks)

	d.SetId(acceptRisks.ID)

	return resourceAcceptRiskRead(ctx, d, m)
}

func resourceAcceptRiskRead(ctx context.Context, d *schema.ResourceData, m interface{}) (diags diag.Diagnostics) {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	acceptRisk, err := sc.GetAcceptRiskRule(d.Id())
	if err != nil {
		diags = append(diags, handleNotFoundError(d, err)...)
		return
	}

	Logf(logDebug, "response: %+v", acceptRisk)

	d.Set("host_type", acceptRisk.HostType)
	d.Set("host_value", acceptRisk.HostValue)

	d.Set("port", acceptRisk.Port)
	d.Set("protocol", acceptRisk.Protocol)
	d.Set("comments", acceptRisk.Comments)
	d.Set("repository_id", acceptRisk.Repository.ID)
	d.SetId(acceptRisk.ID)

	d.Set("plugin_id", acceptRisk.Plugin.ID)

	expiration := acceptRisk.Expires
	expInt, parseErr := strconv.ParseInt(expiration, 10, 64)
	if parseErr != nil {
		diags = append(diags, diag.FromErr(parseErr)...)
		return
	}

	if expInt != -1 {
		expTime := time.Unix(expInt, 0)
		expString := expTime.Format(timeLayout)
		d.Set("expiration", expString)
		if expTime.Before(time.Now()) {
			diags = append(diags,
				diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  fmt.Sprintf("Time %s for expiration is in the past; expired rules are deleted automatically by tenable.sc.", expString),
				},
			)
		}
	} else {
		d.Set("expiration", "-1")
	}

	return
}

func resourceAcceptRiskUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	deleteError := resourceAcceptRiskDelete(ctx, d, m)
	if deleteError != nil {
		return deleteError
	}

	createError := resourceAcceptRiskCreate(ctx, d, m)
	if createError != nil {
		return createError
	}

	return resourceAcceptRiskRead(ctx, d, m)
}

func resourceAcceptRiskDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	acceptID := d.Id()

	err := sc.DeleteAcceptRiskRule(acceptID)
	if err != nil {
		return handleNotFoundError(d, err)
	}

	return nil
}

func validateAcceptRiskExpirationInFuture(expiration any, path cty.Path) (diags diag.Diagnostics) {
	if expiration == nil {
		return
	}

	expAsString, ok := expiration.(string)
	if !ok {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       fmt.Sprintf("expiration value %v is not expected string type", expiration),
			AttributePath: path,
		})
		return
	}
	expTime, err := time.Parse(timeLayout, expAsString)
	if err != nil {
		diags = append(diags, diag.FromErr(err)...)
	}

	if expTime.Before(time.Now()) {
		diags = append(diags,
			diag.Diagnostic{
				Severity:      diag.Warning,
				Summary:       fmt.Sprintf("Time %s for expiration is in the past; expired rules are deleted automatically by tenable.sc.", expAsString),
				AttributePath: path,
			},
		)
	}
	return
}
