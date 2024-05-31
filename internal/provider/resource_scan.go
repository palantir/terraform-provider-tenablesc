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

// ResourceScan Initialize the Accept Risk Resource
func ResourceScan() *schema.Resource {
	return &schema.Resource{
		Description:   descriptionResourceScan,
		CreateContext: resourceScanCreate,
		ReadContext:   resourceScanRead,
		UpdateContext: resourceScanUpdate,
		DeleteContext: resourceScanDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  descriptionDefaultDescriptionValue,
			},
			"repository_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"policy_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"scan_virtual_hosts": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  "false",
			},
			"dhcp_tracking": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  "true",
			},
			"timeout_action": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "import",
			},
			"max_scan_time": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "unlimited",
			},
			"inactivity_timeout": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "3600",
			},
			"ips_and_names": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"asset_ids": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"credential_ids": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"schedule_start": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"schedule_repeat_rule": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
		},
	}
}

func resourceScanCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")

	sc := m.(*tenablesc.Client)

	scan, err := sc.CreateScan(buildScanInputs(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(string(scan.ID))

	return resourceScanRead(ctx, d, m)
}

func resourceScanRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	scan, err := sc.GetScan(d.Id())
	if err != nil {
		return handleNotFoundError(d, err)
	}

	Logf(logDebug, "response: %+v", scan)

	d.SetId(string(scan.ID))
	d.Set("name", scan.Name)
	d.Set("description", scan.Description)
	d.Set("repository_id", scan.Repository.ID)
	d.Set("policy_id", scan.Policy.ID)
	d.Set("scan_virtual_hosts", scan.ScanningVirtualHosts.AsBool())
	d.Set("dhcp_tracking", scan.DHCPTracking.AsBool())
	d.Set("timeout_action", scan.TimeoutAction)
	d.Set("max_scan_time", scan.MaxScanTime)
	d.Set("ips_and_names", scan.IPList)
	d.Set("schedule_start", scan.Schedule.Start)
	d.Set("schedule_repeat_rule", scan.Schedule.RepeatRule)

	var assetIDs []string
	for _, asset := range scan.Assets {
		assetIDs = append(assetIDs, string(asset.ID))
	}
	d.Set("asset_ids", assetIDs)

	var credentialIDs []string
	for _, credential := range scan.Credentials {
		credentialIDs = append(credentialIDs, string(credential.ID))
	}
	d.Set("credential_ids", credentialIDs)

	return nil
}

func resourceScanUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	scan, err := sc.UpdateScan(buildScanInputs(d))
	if err != nil {
		return diag.FromErr(err)
	}

	Logf(logDebug, "response: %+v", scan)

	return resourceScanRead(ctx, d, m)
}

func resourceScanDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	err := sc.DeleteScan(d.Id())
	if err != nil {
		return handleNotFoundError(d, err)
	}

	return nil
}

func buildScanInputs(d *schema.ResourceData) *tenablesc.Scan {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	repositoryID := d.Get("repository_id").(string)
	policyID := d.Get("policy_id").(string)
	scanningVirtualHosts := d.Get("scan_virtual_hosts").(bool)
	dhcpTracking := d.Get("dhcp_tracking").(bool)
	timeoutAction := d.Get("timeout_action").(string)
	maxScanTime := d.Get("max_scan_time").(string)
	inactivityTimeout := d.Get("inactivity_timeout").(string)
	ipsNames := d.Get("ips_and_names").(string)
	assetIDs := d.Get("asset_ids").([]interface{})
	credentialIDs := d.Get("credential_ids").([]interface{})
	scheduleStart := d.Get("schedule_start").(string)
	scheduleRepeatRule := d.Get("schedule_repeat_rule").(string)

	scInput := &tenablesc.Scan{
		BaseInfo: tenablesc.BaseInfo{
			ID:          tenablesc.ProbablyString(d.Id()),
			Name:        name,
			Description: description,
		},

		Type:                 "policy",
		Policy:               &tenablesc.BaseInfo{ID: tenablesc.ProbablyString(policyID)},
		Repository:           &tenablesc.BaseInfo{ID: tenablesc.ProbablyString(repositoryID)},
		DHCPTracking:         tenablesc.ToFakeBool(dhcpTracking),
		ScanningVirtualHosts: tenablesc.ToFakeBool(scanningVirtualHosts),
		TimeoutAction:        timeoutAction,
		MaxScanTime:          maxScanTime,
		InactivityTimeout:    inactivityTimeout,
		IPList:               ipsNames,
	}

	assetBundle := bundleIDs(assetIDs)
	if len(assetBundle) > 0 {
		scInput.Assets = assetBundle
	}
	credBundle := bundleIDs(credentialIDs)
	if len(credBundle) > 0 {
		scInput.Credentials = credBundle
	}

	scheduleType := "ical"
	if scheduleStart == "" {
		scheduleType = "template"
	}

	scInput.Schedule = &tenablesc.ScanSchedule{
		Type:       scheduleType,
		Start:      scheduleStart,
		RepeatRule: scheduleRepeatRule,
	}

	return scInput
}

func bundleIDs(ids []interface{}) []tenablesc.BaseInfo {
	var processedIDs []tenablesc.BaseInfo
	for _, id := range ids {
		obj := tenablesc.BaseInfo{ID: tenablesc.ProbablyString(id.(string))}
		processedIDs = append(processedIDs, obj)
	}

	return processedIDs
}
