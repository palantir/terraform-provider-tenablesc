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
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/palantir/tenablesc-client/tenablesc"
)

func DataSourceAssets() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAssetsRead,
		Description: descriptionDataSourceAssets,
		Schema: map[string]*schema.Schema{
			"assets": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: fmt.Sprintf(descriptionMapIDToNameTemplate, "asset", "asset"),
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"name_filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     ".*",
				Description: fmt.Sprintf(descriptionRegexpNameFilterTemplate, "asset"),
			},
		},
	}
}

func dataSourceAssetsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sc := m.(*tenablesc.Client)

	Logf(logDebug, "looking up all assets")

	allAssets, err := sc.GetAllAssets()
	if err != nil {
		return diag.FromErr(err)
	}

	Logf(logDebug, "response: %+v", allAssets)

	assets := make(map[string]interface{})

	nameFilter := d.Get("name_filter").(string)

	d.SetId(fmt.Sprintf("assets:%s", nameFilter))

	var nameRE *regexp.Regexp

	if len(nameFilter) == 0 {
		return diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("filter is empty string, will return no entries."),
		}}
	}

	nameRE, err = regexp.Compile("^" + nameFilter + "$")
	if err != nil {
		return diag.FromErr(err)
	}

	for _, asset := range allAssets {
		if nameRE.MatchString(asset.Name) {
			assets[string(asset.ID)] = asset.Name
		}
	}

	Logf(logDebug, "Result set: %v", assets)

	if len(assets) == 0 {
		return diag.Errorf("no assets matching filter '^%s$'", nameFilter)
	}

	d.Set("assets", assets)

	return nil
}
