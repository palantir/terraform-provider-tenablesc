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
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/palantir/tenablesc-client/tenablesc"
)

// ResourceAgentScan Initialize the Resource Agent Scan
func ResourceAgentScan() *schema.Resource {
	return &schema.Resource{
		Description:   descriptionResourceAgentScan,
		CreateContext: resourceAgentScanCreate,
		ReadContext:   resourceAgentScanRead,
		UpdateContext: resourceAgentScanUpdate,
		DeleteContext: resourceAgentScanDelete,

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
				Optional: true,
				Default:  "-1",
			},
			"scan_window": {
				Type:     schema.TypeString,
				Required: true,
			},
			"nessus_manager_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"agent_group_ids": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			"email_on_launch": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  "false",
			},

			"email_on_finish": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  "false",
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

func resourceAgentScanCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	sc := m.(*tenablesc.Client)

	agentScan, err := sc.CreateAgentScan(buildAgentScanInput(d))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create agent agent scan: %w", err))
	}

	d.SetId(string(agentScan.ID))

	Logf(logDebug, "agentScan: %+v", agentScan)

	return resourceAgentScanRead(ctx, d, m)
}

func resourceAgentScanRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	sc := m.(*tenablesc.Client)

	agentScan, err := sc.GetAgentScan(d.Id())
	if err != nil {
		return handleNotFoundError(d, err)
	}

	Logf(logDebug, "response: %+v", agentScan)

	d.Set("name", agentScan.Name)
	d.Set("description", agentScan.Description)
	d.Set("repository_id", agentScan.Repository.ID)
	d.Set("policy_id", agentScan.Policy.ID)
	d.Set("scan_window", agentScan.ScanWindow)
	d.Set("nessus_manager_id", agentScan.NessusManager.ID)
	d.Set("email_on_launch", agentScan.EmailOnLaunch.AsBool())
	d.Set("email_on_finish", agentScan.EmailOnFinish.AsBool())
	d.Set("schedule_start", agentScan.Schedule.Start)
	d.Set("schedule_repeat_rule", agentScan.Schedule.RepeatRule)

	var agentGroupIDs []string
	for _, agentGroup := range agentScan.AgentGroups {
		agentGroupIDs = append(agentGroupIDs, strconv.Itoa(agentGroup.RemoteID))
	}

	d.Set("agent_group_ids", agentGroupIDs)
	d.SetId(string(agentScan.ID))
	return nil
}

func resourceAgentScanUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	sc := m.(*tenablesc.Client)

	scan, err := sc.UpdateAgentScan(buildAgentScanInput(d))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to update agent scan: %w", err))
	}

	Logf(logDebug, "response: %+v", scan)

	return resourceAgentScanRead(ctx, d, m)
}

func resourceAgentScanDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	sc := m.(*tenablesc.Client)

	err := sc.DeleteAgentScan(d.Id())
	if err != nil {
		return handleNotFoundError(d, err)
	}

	return nil
}

func buildAgentScanInput(d *schema.ResourceData) *tenablesc.AgentScan {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	repositoryID := d.Get("repository_id").(string)
	policyID := d.Get("policy_id").(string)
	scanWindow := d.Get("scan_window").(string)
	nessusMangerID := d.Get("nessus_manager_id").(string)
	agentGroupIDs := d.Get("agent_group_ids").([]any)
	emailOnLaunch := d.Get("email_on_launch").(bool)
	emailOnFinish := d.Get("email_on_finish").(bool)
	scheduleStart := d.Get("schedule_start").(string)
	scheduleRepeatRule := d.Get("schedule_repeat_rule").(string)

	scInput := &tenablesc.AgentScan{
		ID:          tenablesc.ProbablyString(d.Id()),
		Name:        name,
		Description: description,
		Policy:      &tenablesc.BaseInfo{ID: tenablesc.ProbablyString(policyID)},
		Type:        "policy",
		Repository:  tenablesc.BaseInfo{ID: tenablesc.ProbablyString(repositoryID)},
		NessusManager: tenablesc.BaseInfo{
			ID:          tenablesc.ProbablyString(nessusMangerID),
			Name:        name,
			Description: description,
		},
		ScanWindow:    tenablesc.ProbablyString(scanWindow),
		AgentGroups:   toAgentGroups(agentGroupIDs),
		EmailOnLaunch: tenablesc.ToFakeBool(emailOnLaunch),
		EmailOnFinish: tenablesc.ToFakeBool(emailOnFinish),
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
func toAgentGroups(agentGroupsInterface []any) []tenablesc.AgentGroup {
	var agentGroups []tenablesc.AgentGroup
	for _, item := range agentGroupsInterface {
		agentGroups = append(agentGroups, tenablesc.AgentGroup{
			ID: tenablesc.ProbablyString(item.(string))})
	}

	return agentGroups
}
