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
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/palantir/tenablesc-client/tenablesc"
)

// ResourceScanPolicy Initialize the Accept Risk Resource
func ResourceScanPolicy() *schema.Resource {
	return &schema.Resource{
		Description:   descriptionResourceScanPolicy,
		CreateContext: resourceScanPolicyCreate,
		ReadContext:   resourceScanPolicyRead,
		UpdateContext: resourceScanPolicyUpdate,
		DeleteContext: resourceScanPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: descriptionScanPolicyName,
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: descriptionScanPolicyDescription,
				Optional:    true,
				Default:     descriptionDefaultDescriptionValue,
			},
			"policy_template_id": {
				Type:        schema.TypeString,
				Description: descriptionScanPolicyTemplateID,
				Required:    true,
			},
			"audit_file_id": {
				Type:        schema.TypeString,
				Description: descriptionAuditFileID,
				Optional:    true,
				Default:     "",
			},
			"preferences": {
				Type:        schema.TypeMap,
				Description: descriptionScanPolicyPreferences,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"families": {
				Type:        schema.TypeSet,
				Description: descriptionScanPolicyFamilies,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"tag": {
				Type:        schema.TypeString,
				Description: descriptionScanPolicyTag,
				Optional:    true,
				Default:     "",
			},
		},
	}
}

func resourceScanPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")

	sc := m.(*tenablesc.Client)

	inputs, err := buildScanPolicyInputs(d)
	if err != nil {
		return diag.FromErr(err)
	}
	policy, err := sc.CreateScanPolicy(inputs)
	if err != nil {
		return diag.FromErr(err)
	}

	Logf(logDebug, "response: %+v", policy)

	d.SetId(string(policy.ID))

	return resourceScanPolicyRead(ctx, d, m)
}

func resourceScanPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")

	sc := m.(*tenablesc.Client)

	policy, err := sc.GetScanPolicy(d.Id())
	if err != nil {
		return handleNotFoundError(d, err)
	}

	Logf(logDebug, "response: %+v", policy)

	d.Set("name", policy.Name)
	d.Set("description", policy.Description)
	if policy.PolicyTemplate != nil {
		d.Set("policy_template_id", policy.PolicyTemplate.ID)
	}

	upstreamPolicyPreferencesMap, ok := policy.Preferences.(map[string]any)
	if !ok {
		return diag.Errorf("Could not render current policy preferences (type %T) as a map[string]any", policy.Preferences)
	}

	marshalledUpstreamPreferences, err := marshalPreferenceMap(upstreamPolicyPreferencesMap)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("preferences", marshalledUpstreamPreferences)

	var tfFamilies []string
	for _, family := range policy.Families {
		tfFamilies = append(tfFamilies, family.ID)
	}
	d.Set("families", tfFamilies)

	if len(policy.AuditFiles) > 0 {
		d.Set("audit_file_id", policy.AuditFiles[0].ID)
	} else {
		d.Set("audit_file_id", "")
	}

	d.SetId(string(policy.ID))

	return nil
}

func resourceScanPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")

	sc := m.(*tenablesc.Client)

	inputs, err := buildScanPolicyInputs(d)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = sc.UpdateScanPolicy(inputs)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceScanPolicyRead(ctx, d, m)
}

func resourceScanPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")

	sc := m.(*tenablesc.Client)

	err := sc.DeleteScanPolicy(d.Id())
	if err != nil {
		return handleNotFoundError(d, err)
	}

	return nil
}

// marshalPreferenceMap renders a map that contains actual data structures in to a map[string]string.
// returning map[string]interface{} because TF underpinnings expect it.
func marshalPreferenceMap(m map[string]interface{}) (map[string]interface{}, error) {

	marshalled := make(map[string]interface{})

	for k, v := range m {
		if vs, ok := v.(string); ok {
			marshalled[k] = vs
			continue
		}
		marshalledValue, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		marshalled[k] = string(marshalledValue)
	}
	Logf(logDebug, "Marshalled Preference Map: %+v", marshalled)
	return marshalled, nil
}

// unmarshalPreferencesMap renders a map containing jsonified array values into their native structures.
func unmarshalPreferencesMap(m map[string]interface{}) (map[string]interface{}, error) {

	unmarshalled := make(map[string]interface{})

	for k, v := range m {
		var unmarshalledArray []string
		err := json.Unmarshal([]byte(v.(string)), &unmarshalledArray)
		if err == nil {
			unmarshalled[k] = unmarshalledArray
			continue
		}
		unmarshalled[k] = v.(string)
	}

	Logf(logDebug, "Unmarshalled Preference Map: %+v", unmarshalled)
	return unmarshalled, nil
}

func buildScanPolicyInputs(d *schema.ResourceData) (*tenablesc.ScanPolicy, error) {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	policyTemplateID := d.Get("policy_template_id").(string)
	auditFileID := d.Get("audit_file_id").(string)
	families := d.Get("families").(*schema.Set).List()
	tag := d.Get("tag").(string)

	// To identify what to remove from prefs, look at what's in state vs what we're
	// replacing it with; anything in state but not in config should be removed.
	oldPreferences, newPreferences := d.GetChange("preferences")
	oldPreferencesMap, ok := oldPreferences.(map[string]any)
	if !ok {
		oldPreferencesMap = make(map[string]any)
	}
	newPreferencesMap, ok := newPreferences.(map[string]any)
	if !ok {
		newPreferencesMap = make(map[string]any)
	}

	var prefsToRemove []string

	for oldPrefName := range oldPreferencesMap {
		if _, exists := newPreferencesMap[oldPrefName]; !exists {
			removePref := oldPrefName
			prefsToRemove = append(prefsToRemove, removePref)
		}
	}

	spInput := &tenablesc.ScanPolicy{
		BaseInfo: tenablesc.BaseInfo{
			ID:          tenablesc.ProbablyString(d.Id()),
			Name:        name,
			Description: description,
		},
		Tags:           tag,
		PolicyTemplate: &tenablesc.BaseInfo{ID: tenablesc.ProbablyString(policyTemplateID)},
	}

	prefStringMap, err := unmarshalPreferencesMap(newPreferencesMap)
	if err != nil {
		return nil, err
	}
	spInput.Preferences = prefStringMap

	spInput.RemovePrefs = prefsToRemove

	var famInput []tenablesc.ScanPolicyFamilies
	for _, family := range families {
		famInput = append(famInput, tenablesc.ScanPolicyFamilies{ID: family.(string)})
	}
	if len(famInput) > 0 {
		spInput.Families = famInput
	}

	if auditFileID != "" {
		spInput.AuditFiles = []tenablesc.BaseInfo{{ID: tenablesc.ProbablyString(auditFileID)}}
	}

	return spInput, nil
}
