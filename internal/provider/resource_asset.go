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

// ResourceAsset Initialize the Accept Risk Resource
func ResourceAsset() *schema.Resource {
	return &schema.Resource{
		Description:   descriptionResourceAsset,
		CreateContext: resourceAssetCreate,
		ReadContext:   resourceAssetRead,
		UpdateContext: resourceAssetUpdate,
		DeleteContext: resourceAssetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: descriptionAssetName,
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: descriptionAssetDescription,
				Optional:    true,
				Default:     descriptionDefaultDescriptionValue,
			},
			"type": {
				Type:        schema.TypeString,
				Description: descriptionAssetType,
				Required:    true,
				ForceNew:    true,
			},
			"values": {
				Type:        schema.TypeSet,
				Description: descriptionAssetValues,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceAssetCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	response, err := sc.CreateAsset(buildAssetInput(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(string(response.ID))

	Logf(logDebug, "response: %+v", response)

	return resourceAssetRead(ctx, d, m)
}

func resourceAssetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	assetResponse, err := sc.GetAsset(d.Id())
	if err != nil {
		return handleNotFoundError(d, err)
	}

	Logf(logDebug, "response: %+v", assetResponse)

	d.Set("name", assetResponse.Name)
	d.Set("description", assetResponse.Description)
	d.Set("type", assetResponse.Type)
	switch assetResponse.Type {
	case "dnsname":
		d.Set("values", assetResponse.DefinedDNSNames)
	case "static":
		d.Set("values", assetResponse.DefinedIPs)
	}
	d.SetId(string(assetResponse.ID))

	return nil
}

func resourceAssetUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	_, err := sc.UpdateAsset(buildAssetInput(d))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceAssetRead(ctx, d, m)
}

func resourceAssetDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	assetID := d.Id()

	err := sc.DeleteAsset(assetID)
	if err != nil {
		return handleNotFoundError(d, err)
	}

	return nil
}

func buildAssetInput(d *schema.ResourceData) *tenablesc.Asset {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	assetType := d.Get("type").(string)
	values := d.Get("values").(*schema.Set)

	assetInput := &tenablesc.Asset{
		BaseInfo: tenablesc.BaseInfo{
			ID:          tenablesc.ProbablyString(d.Id()),
			Name:        name,
			Description: description,
		},
		Type: assetType,
	}

	var assetList []string
	for _, v := range values.List() {
		assetList = append(assetList, v.(string))
	}

	switch assetType {
	case "dnsname":
		assetInput.DefinedDNSNames = assetList
	case "static":
		assetInput.DefinedIPs = assetList
	}

	return assetInput
}
