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
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/palantir/tenablesc-client/tenablesc"
)

func DataSourceAsset() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAssetRead,
		Description: descriptionDataSourceAsset,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf(descriptionDataSourceNameFindTemplate, "asset"),
			},
			"defined_dns_names": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: descriptionAssetDefinedDNSNames,
			},
			"defined_ips": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: descriptionAssetDefinedIPs,
			},
		},
	}
}

func dataSourceAssetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sc := m.(*tenablesc.Client)
	assetName := d.Get("name").(string)

	Logf(logDebug, "looking up %s", assetName)

	allAssets, err := sc.GetAllAssets()
	if err != nil {
		return diag.FromErr(err)
	}

	for _, asset := range allAssets {
		Logf(logTrace, "comparing asset: %+v", *asset)
		if strings.Compare(asset.Name, assetName) == 0 {
			d.SetId(string(asset.ID))
			d.Set("definedDNSNames", asset.DefinedDNSNames)
			d.Set("definedIPs", asset.DefinedIPs)
			return nil
		}
	}

	return diag.Errorf("No asset with name [%s] found", assetName)
}
