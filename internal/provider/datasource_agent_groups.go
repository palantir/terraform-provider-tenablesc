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
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/palantir/tenablesc-client/tenablesc"
)

func DataSourceAgentGroups() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAgentGroupsRead,
		Description: descriptionDataSourceAgentGroups,
		Schema: map[string]*schema.Schema{
			"agent_groups": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: fmt.Sprintf(descriptionMapIDToNameTemplate, "agent group", "agent group"),
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"agent_scanner_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "specify a remote agent scanner ID",
			},
			"name_filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     ".*",
				Description: fmt.Sprintf(descriptionRegexpNameFilterTemplate, "agent group"),
			},
		},
	}
}

func dataSourceAgentGroupsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sc := m.(*tenablesc.Client)

	Logf(logDebug, "looking up all agent groups")

	agentScannerID := d.Get("agent_scanner_id").(string)
	agentGroups, err := sc.GetAgentGroupsForScanner(agentScannerID)
	errors.Is(err, tenablesc.NotFoundError{})
	nfe := tenablesc.NotFoundError{}
	if errors.As(err, &nfe) {
		return nil
	}
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get agent groups for scanner %s: %w", agentScannerID, err))
	}

	Logf(logDebug, "response: %+v", agentGroups)

	nameFilter := d.Get("name_filter").(string)

	if len(nameFilter) == 0 {
		return diag.Diagnostics{{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("filter is empty string, will return no entries."),
		}}
	}

	nameRE, err := regexp.Compile("^" + nameFilter + "$")
	if err != nil {
		return diag.FromErr(err)
	}

	filteredAgentGroups := make(map[string]interface{})
	for _, agentGroup := range agentGroups {
		if nameRE.MatchString(agentGroup.Name) {
			filteredAgentGroups[string(agentGroup.ID)] = agentGroup.Name
		}
	}

	Logf(logDebug, "Result set: %v", filteredAgentGroups)

	if len(filteredAgentGroups) == 0 {
		return diag.Errorf("no agent groups matching filter '^%s$'", nameFilter)
	}

	d.SetId(fmt.Sprintf("agent_groups:%s:%s", nameFilter, agentScannerID))
	d.Set("agent_groups", filteredAgentGroups)

	return nil
}
