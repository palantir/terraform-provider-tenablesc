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

// Provider implement the SC Provider
func Provider() *schema.Provider {
	return &schema.Provider{
		ConfigureContextFunc: configureProvider,
		ResourcesMap: map[string]*schema.Resource{
			"tenablesc_accept_risk":                         ResourceAcceptRisk(),
			"tenablesc_asset":                               ResourceAsset(),
			"tenablesc_auditfile":                           ResourceAuditFile(),
			"tenablesc_organization":                        ResourceOrganization(),
			"tenablesc_recast_risk":                         ResourceRecastRisk(),
			"tenablesc_repository":                          ResourceRepository(),
			"tenablesc_scan_policy":                         ResourceScanPolicy(),
			"tenablesc_scan":                                ResourceScan(),
			"tenablesc_scan_zone":                           ResourceScanZone(),
			"tenablesc_repository_organization_association": ResourceRepositoryOrganizationAssociation(),
			"tenablesc_organization_scan_zone_association":  ResourceOrganizationScanZoneAssociation(),
			"tenablesc_role":                                ResourceRole(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"tenablesc_plugin":               DataSourcePlugin(),
			"tenablesc_repository":           DataSourceRepository(),
			"tenablesc_repositories":         DataSourceRepositories(),
			"tenablesc_asset":                DataSourceAsset(),
			"tenablesc_assets":               DataSourceAssets(),
			"tenablesc_scan_policy_template": DataSourceScanPolicyTemplate(),
			"tenablesc_credential":           DataSourceCredential(),
		},
		Schema: map[string]*schema.Schema{
			"uri": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("TENABLESC_URI", nil),
				Description: "URI of the REST API endpoint. This serves as the base of all requests.",
			},
			"access_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("TENABLESC_ACCESS_KEY", nil),
				Description: "SC Access Key to use",
			},
			"secret_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("TENABLESC_SECRET_KEY", nil),
				Description: "SC Secret Key to use",
			},
		},
	}
}

func configureProvider(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	accessKey := d.Get("access_key").(string)
	secretKey := d.Get("secret_key").(string)
	scURI := d.Get("uri").(string)

	client := tenablesc.NewClient(scURI).SetAPIKey(accessKey, secretKey)

	currentUser, err := client.GetCurrentUser()
	if err != nil {
		return nil, diag.FromErr(err)
	}

	Logf(logDebug, "Configured provider with user %+v", *currentUser)

	return client, nil
}
