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

func ResourceOrganizationScanZoneAssociation() *schema.Resource {
	return &schema.Resource{
		Description:   descriptionResourceOrganizationScanZoneAssociation,
		CreateContext: resourceOrganizationScanZoneAssociationCreateOrUpdate,
		ReadContext:   resourceOrganizationScanZoneAssociationRead,
		UpdateContext: resourceOrganizationScanZoneAssociationCreateOrUpdate,
		DeleteContext: resourceOrganizationScanZoneAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:        schema.TypeString,
				Description: descriptionOrganizationID,
				Required:    true,
			},
			"scan_zone_ids": {
				Type:        schema.TypeSet,
				Description: descriptionOrganizationScanZoneIDs,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Required:    true,
			},
		},
	}
}

func resourceOrganizationScanZoneAssociationCreateOrUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	d.SetId(d.Get("organization_id").(string))

	orgAssociationInput, err := buildOrganizationScanZoneAssociationInputs(d)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = sc.UpdateOrganization(orgAssociationInput)
	if err != nil {
		return handleNotFoundError(d, err)
	}

	return resourceOrganizationScanZoneAssociationRead(ctx, d, m)
}

func resourceOrganizationScanZoneAssociationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	id := d.Get("organization_id").(string)

	org, err := sc.GetOrganization(id)
	if err != nil {
		return handleNotFoundError(d, err)
	}

	Logf(logDebug, "response: %+v", org)

	idInt, err := strconv.ParseInt(string(org.ID), 10, 32)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("organization_id", int(idInt))
	d.SetId(string(org.ID))

	var zoneList []int
	for _, zone := range org.Zones {
		zoneID, err := strconv.ParseInt(string(zone.ID), 10, 32)
		if err != nil {
			return diag.FromErr(err)
		}
		zoneList = append(zoneList, int(zoneID))
	}
	d.Set("scan_zone_ids", zoneList)

	return nil
}

func resourceOrganizationScanZoneAssociationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	id := d.Id()

	orgAssociation := &tenablesc.Organization{
		BaseInfo: tenablesc.BaseInfo{ID: tenablesc.ProbablyString(id)},
		Zones:    []tenablesc.BaseInfo{},
	}

	_, err := sc.UpdateOrganization(orgAssociation)
	if err != nil {
		return handleNotFoundError(d, err)
	}
	return nil
}

func buildOrganizationScanZoneAssociationInputs(d *schema.ResourceData) (*tenablesc.Organization, error) {

	orgID := d.Id()
	zones := d.Get("scan_zone_ids").(*schema.Set)

	orgAssociation := &tenablesc.Organization{
		BaseInfo: tenablesc.BaseInfo{ID: tenablesc.ProbablyString(orgID)},
		Zones:    []tenablesc.BaseInfo{},
	}

	for _, i := range zones.List() {
		orgAssociation.Zones = append(orgAssociation.Zones,
			tenablesc.BaseInfo{ID: tenablesc.ProbablyString(strconv.Itoa(i.(int)))},
		)
	}

	return orgAssociation, nil
}
