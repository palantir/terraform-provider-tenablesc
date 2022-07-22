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

func DataSourcePlugin() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePluginRead,
		Description: descriptionDataSourcePlugin,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: descriptionPluginName,
			},
		},
	}
}

func dataSourcePluginRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	pluginName := d.Get("name").(string)

	Logf(logDebug, "looking up %s", pluginName)

	sc := m.(*tenablesc.Client)

	pluginResponse, err := sc.GetPluginsByName(pluginName)

	if err != nil {
		return diag.Errorf("plugin datasource lookup failed: %v", err)
	}

	if len(pluginResponse) == 0 {
		return diag.Errorf("no Plugin found with name: %s", pluginName)
	}

	if len(pluginResponse) > 1 {
		return diag.Errorf("got ambiguous result, %d plugins for name %s", len(pluginResponse), pluginName)
	}

	pluginID := pluginResponse[0].ID
	d.SetId(string(pluginID))

	return nil
}
