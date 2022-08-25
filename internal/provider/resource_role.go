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
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/palantir/tenablesc-client/tenablesc"
)

// maps terraform attribute names to role struct names
// Because the permissions are numerous and identical, handling by hand seemed like a legibility nightmare.
// See getRolePermission and setRolePermission down below for tools to handle the mapped access with
// reflection.
const rolePermissionPrefix = "perm_"

var rolePermissions = map[string]string{
	"manage_groups":              "PermManageGroups",
	"manage_roles":               "PermManageRoles",
	"manage_images":              "PermManageImages",
	"manage_group_relationships": "PermManageGroupRelationships",
	"manage_blackout_windows":    "PermManageBlackoutWindows",
	"manage_attribute_sets":      "PermManageAttributeSets",
	"create_tickets":             "PermCreateTickets",
	"create_audit_files":         "PermCreateAuditFiles",
	"create_ldap_assets":         "PermCreateLDAPAssets",
	"create_policies":            "PermCreatePolicies",
	"purge_tickets":              "PermPurgeTickets",
	"purge_scan_results":         "PermPurgeScanResults",
	"purge_report_results":       "PermPurgeReportResults",
	"scan":                       "PermScan",
	"agents_scan":                "PermAgentsScan",
	"share_objects":              "PermShareObjects",
	"update_feeds":               "PermUpdateFeeds",
	"upload_nessus_results":      "PermUploadNessusResults",
	"view_org_logs":              "PermViewOrgLogs",
	"manage_accept_risk_rules":   "PermManageAcceptRiskRules",
	"manage_recast_risk_rules":   "PermManageRecastRiskRules",
}

// ResourceRole provides the User Role permission mapping.
func ResourceRole() *schema.Resource {
	resource := &schema.Resource{
		Description:   descriptionResourceRole,
		CreateContext: resourceRoleCreate,
		ReadContext:   resourceRoleRead,
		UpdateContext: resourceRoleUpdate,
		DeleteContext: resourceRoleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: descriptionRoleName,
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: descriptionRoleDescription,
				Optional:    true,
				Default:     descriptionDefaultDescriptionValue,
			},
		},
	}

	for t := range rolePermissions {
		resource.Schema[fmt.Sprintf("%s%s", rolePermissionPrefix, t)] = &schema.Schema{
			Type:        schema.TypeBool,
			Optional:    true,
			Description: fmt.Sprintf("Set permission flag %s", rolePermissions[t]),
			Default:     false,
		}
	}

	return resource
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")

	sc := m.(*tenablesc.Client)

	roleInput, err := buildRoleInput(d)
	if err != nil {
		return diag.FromErr(err)
	}

	response, err := sc.CreateRole(roleInput)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(string(response.ID))

	return resourceRoleRead(ctx, d, m)
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	roleResponse, err := sc.GetRole(d.Id())
	if err != nil {
		return handleNotFoundError(d, err)
	}

	Logf(logDebug, "response: %+v", roleResponse)

	d.Set("name", roleResponse.Name)
	d.Set("description", roleResponse.Description)
	d.SetId(string(roleResponse.ID))

	for k, v := range rolePermissions {
		b, err := getRolePermission(roleResponse, v)
		if err != nil {
			return diag.FromErr(err)
		}
		d.Set(k, b)
	}

	return nil
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")

	sc := m.(*tenablesc.Client)

	roleInput, err := buildRoleInput(d)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = sc.UpdateRole(roleInput)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceRoleRead(ctx, d, m)
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")

	sc := m.(*tenablesc.Client)

	err := sc.DeleteRole(d.Id())
	if err != nil {
		return handleNotFoundError(d, err)
	}

	return nil
}

func buildRoleInput(d *schema.ResourceData) (*tenablesc.Role, error) {
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	roleInput := &tenablesc.Role{
		BaseInfo: tenablesc.BaseInfo{
			ID:          tenablesc.ProbablyString(d.Id()),
			Name:        name,
			Description: description,
		},
	}

	for k, v := range rolePermissions {
		if b, ok := d.GetOk(k); ok {
			if err := setRolePermission(roleInput, v, b.(bool)); err != nil {
				return nil, err
			}
		}
	}

	return roleInput, nil
}

func setRolePermission(role *tenablesc.Role, name string, value bool) error {
	ps := reflect.ValueOf(role)
	s := ps.Elem()
	if s.Kind() != reflect.Struct {
		return errors.New("unable to get struct for role")
	}
	f := s.FieldByName(name)

	if !f.IsValid() || !f.CanSet() {
		return fmt.Errorf("cannot set invalid field %s", name)
	}

	if !(f.Kind() == reflect.String) {
		return fmt.Errorf("can't set non-string field %s", name)
	}

	fakeBoolValue := tenablesc.ToFakeBool(value)

	f.SetString(string(fakeBoolValue))

	return nil
}

func getRolePermission(role *tenablesc.Role, name string) (bool, error) {
	ps := reflect.ValueOf(role)
	s := ps.Elem()
	if s.Kind() != reflect.Struct {
		return false, errors.New("unable to get struct for role")
	}
	f := s.FieldByName(name)

	if !f.IsValid() {
		return false, fmt.Errorf("cannot get invalid field %s", name)
	}

	if !(f.Kind() == reflect.String) {
		return false, fmt.Errorf("can't set non-string field %s", name)
	}

	fakeBoolValue := f.String()

	return tenablesc.FakeBool(fakeBoolValue).AsBool(), nil
}
