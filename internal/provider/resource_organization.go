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
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/palantir/tenablesc-client/tenablesc"
)

func ResourceOrganization() *schema.Resource {
	return &schema.Resource{
		Description:   descriptionResourceOrganization,
		CreateContext: resourceOrganizationCreate,
		ReadContext:   resourceOrganizationRead,
		UpdateContext: resourceOrganizationUpdate,
		DeleteContext: resourceOrganizationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: descriptionOrganizationName,
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: descriptionOrganizationDescription,
				Optional:    true,
				Default:     descriptionDefaultDescriptionValue,
			},
			"zone_selection": {
				Type:             schema.TypeString,
				Description:      descriptionOrganizationZoneSelection,
				Optional:         true,
				Default:          "auto_only",
				ValidateDiagFunc: validateZoneSelection,
			},
			"scan_zone_ids": {
				Type:        schema.TypeSet,
				Description: descriptionOrganizationScanZoneIDs,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
			},
			"restricted_ips": {
				Type:        schema.TypeSet,
				Description: descriptionOrganizationRestrictedIPs,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
			},
		},
	}
}

func resourceOrganizationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	orgInputs, diags := buildOrgInputs(d)
	if diags.HasError() {
		return diags
	}
	organization, err := sc.CreateOrganization(orgInputs)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(string(organization.ID))

	return resourceOrganizationRead(ctx, d, m)
}

func resourceOrganizationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	organization, err := sc.GetOrganization(d.Id())

	if err != nil {
		return handleNotFoundError(d, err)
	}

	Logf(logDebug, "response: %+v", organization)

	d.SetId(string(organization.ID))

	d.Set("name", organization.Name)
	d.Set("description", organization.Description)

	d.Set("zone_selection", organization.ZoneSelection)

	if len(organization.RestrictedIPs) > 0 {
		d.Set("restricted_ips", strings.Split(organization.RestrictedIPs, ","))
	}

	return nil
}

func resourceOrganizationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	orgInputs, diags := buildOrgInputs(d)
	if diags.HasError() {
		return diags
	}

	_, err := sc.UpdateOrganization(orgInputs)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceOrganizationRead(ctx, d, m)
}

func resourceOrganizationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	err := sc.DeleteOrganization(d.Id())
	if err != nil {
		return handleNotFoundError(d, err)
	}

	return nil
}

func buildOrgInputs(d *schema.ResourceData) (*tenablesc.Organization, diag.Diagnostics) {
	var diags diag.Diagnostics

	name := d.Get("name").(string)
	description := d.Get("description").(string)

	org := &tenablesc.Organization{
		BaseInfo: tenablesc.BaseInfo{
			ID:          tenablesc.ProbablyString(d.Id()),
			Name:        name,
			Description: description,
		},
		VulnScoringSystem: "CVSSv3", // Hardcoded until we have a reason not to.
	}

	if restrictedIPs, ok := d.GetOk("restricted_ips"); ok {
		ipList := restrictedIPs.(*schema.Set).List()
		ipStringList := make([]string, len(ipList))
		for i, v := range ipList {
			ipStringList[i] = v.(string)
		}
		org.RestrictedIPs = strings.Join(ipStringList, ",")
	}

	org.ZoneSelection = d.Get("zone_selection").(string)
	zoneSet, ok := d.Get("scan_zone_ids").(*schema.Set)
	if !ok {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to parse 'zones' attribute as Set"),
		})
		return nil, diags
	}

	zoneList := zoneSet.List()

	if err := validateZones(org.ZoneSelection, zoneList); err != nil {
		diags = append(diags, diag.FromErr(err)...)
		return nil, diags
	}

	return org, diags

}

func validateZones(selection string, zones []any) error {
	switch selection {
	case "auto_only":
		if len(zones) != 0 {
			return errors.New("zone selection 'auto_only' requires zones not be specified")
		}
	case "locked":
		if len(zones) != 0 {
			return fmt.Errorf("zone selection 'locked' requires a single zone be specified, got %d", len(zones))
		}
	case "selectable", "selectable+auto_restricted":
		if len(zones) < 1 {
			return fmt.Errorf("zone selection '%s' requires at least one zone be specified", selection)
		}
	}

	return nil
}

// no const slices, grrr.
var validZoneSelectors = []string{
	"auto_only",
	"locked",
	"selectable",
	"selectable+auto",
	"selectable+auto_restricted",
}

func validateZoneSelection(i interface{}, path cty.Path) diag.Diagnostics {
	if selection, ok := i.(string); ok {
		for _, v := range validZoneSelectors {
			if selection == v {
				return nil
			}
		}
		return diag.Errorf("%s is not a valid zone selector. Valid selectors are %v", selection, validZoneSelectors)
	}
	return diag.Errorf("could not cast %v to string", i)
}
