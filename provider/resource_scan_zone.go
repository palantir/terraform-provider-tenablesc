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

// ResourceScanZone provides CIDR to scanner and org mappings
func ResourceScanZone() *schema.Resource {
	return &schema.Resource{
		Description:   descriptionResourceScanZone,
		CreateContext: resourceScanZoneCreate,
		ReadContext:   resourceScanZoneRead,
		UpdateContext: resourceScanZoneUpdate,
		DeleteContext: resourceScanZoneDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: descriptionScanZoneName,
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: descriptionScanZoneDescription,
				Optional:    true,
				Default:     descriptionDefaultDescriptionValue,
			},
			"zone_cidrs": {
				Type:        schema.TypeSet,
				Description: descriptionScanZoneCIDRs,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
			},
		},
	}
}

func resourceScanZoneCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	response, err := sc.CreateScanZone(buildScanZoneInput(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(string(response.ID))

	return resourceScanZoneRead(ctx, d, m)
}

func resourceScanZoneRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	scanZoneResponse, err := sc.GetScanZone(d.Id())
	if err != nil {
		return handleNotFoundError(d, err)
	}

	Logf(logDebug, "response: %+v", scanZoneResponse)

	d.Set("name", scanZoneResponse.Name)
	d.Set("description", scanZoneResponse.Description)
	d.SetId(string(scanZoneResponse.ID))

	d.Set("zone_cidrs", scanZoneResponse.IPList)

	return nil
}

func resourceScanZoneUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	_, err := sc.UpdateScanZone(buildScanZoneInput(d))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScanZoneRead(ctx, d, m)
}

func resourceScanZoneDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	err := sc.DeleteScanZone(d.Id())
	if err != nil {
		return handleNotFoundError(d, err)
	}

	return nil
}

func buildScanZoneInput(d *schema.ResourceData) *tenablesc.ScanZone {
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	scanZoneInput := &tenablesc.ScanZone{
		ScanZoneBaseFields: tenablesc.ScanZoneBaseFields{
			BaseInfo: tenablesc.BaseInfo{
				ID:          tenablesc.ProbablyString(d.Id()),
				Name:        name,
				Description: description,
			},
		},
	}

	if zoneCidrs, ok := d.GetOk("zone_cidrs"); ok {
		zcList := zoneCidrs.(*schema.Set).List()

		var zcStrings []string
		for _, zc := range zcList {
			zcStrings = append(zcStrings, zc.(string))
		}

		scanZoneInput.IPList = zcStrings
	}

	return scanZoneInput
}
