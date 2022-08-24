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

// ResourceAuditFile Initialize the Accept Risk Resource
func ResourceAuditFile() *schema.Resource {
	return &schema.Resource{
		Description:   descriptionResourceAuditFile,
		CreateContext: resourceAuditFileCreate,
		ReadContext:   resourceAuditFileRead,
		UpdateContext: resourceAuditFileUpdate,
		DeleteContext: resourceAuditFileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: descriptionAuditFileName,
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: descriptionAuditFileDescription,
				Optional:    true,
				Default:     descriptionDefaultDescriptionValue,
			},
			"content": {
				Type:        schema.TypeString,
				Description: descriptionAuditFileContent,
				Required:    true,
			},
			"sc_filename": {
				Type:        schema.TypeString,
				Description: descriptionAuditFileSCFilename,
				Computed:    true,
			},
		},
	}
}

func resourceAuditFileCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	var err error

	name := d.Get("name").(string)
	content := d.Get("content").(string)

	filename, err := uploadNewAuditFile(sc, name, content)
	if err != nil {
		Logf(logError, "Upload file response: %+v ", err)
		return diag.FromErr(err)
	}

	d.Set("sc_filename", filename)

	response, err := sc.CreateAuditFile(buildAuditFileInput(d))
	if err != nil {
		return diag.FromErr(err)
	}

	Logf(logDebug, "response: %+v", response)

	d.SetId(string(response.ID))

	return resourceAuditFileRead(ctx, d, m)
}

func resourceAuditFileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	auditFile, err := sc.GetAuditFile(d.Id())
	if err != nil {
		return handleNotFoundError(d, err)
	}

	Logf(logDebug, "response: %+v", auditFile)

	d.SetId(string(auditFile.ID))
	d.Set("name", auditFile.Name)
	d.Set("description", auditFile.Description)
	d.Set("sc_filename", auditFile.Filename)

	return nil
}

func uploadNewAuditFile(sc *tenablesc.Client, name, content string) (string, error) {
	file, err := sc.UploadFileFromString(content, name, "auditfile")
	if err != nil {
		return "", err
	}
	return file.Filename, nil
}

func resourceAuditFileUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	oldFilename := d.Get("sc_filename").(string)
	name := d.Get("name").(string)
	content := d.Get("content").(string)
	if d.HasChange("content") {
		filename, err := uploadNewAuditFile(sc, name, content)
		if err != nil {
			return diag.FromErr(err)
		}
		d.Set("sc_filename", filename)
	}

	auditFile, err := sc.UpdateAuditFile(buildAuditFileInput(d))
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("content") && d.HasChange("sc_filename") {
		// Now that we've updated the audit file, including the new content,
		// remove the old file for cleanliness, unless it's the same filename.
		err := sc.DeleteFile(oldFilename)
		if err != nil {
			// If we can't delete the old file, shrug and move on, could be referred to elsewhere
			return diag.Diagnostics{
				{
					Severity: diag.Warning,
					Summary:  "Unable to delete old audit file, may still be in use.",
					Detail:   err.Error(),
				},
			}
		}
	}
	Logf(logDebug, "response: %+v", auditFile)

	return resourceAuditFileRead(ctx, d, m)
}

func resourceAuditFileDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	Logf(logTrace, "start of function")
	sc := m.(*tenablesc.Client)

	err := sc.DeleteAuditFile(d.Id())
	if err != nil {
		return handleNotFoundError(d, err)
	}

	return nil
}

func buildAuditFileInput(d *schema.ResourceData) *tenablesc.AuditFile {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	filename := d.Get("sc_filename").(string)

	afInput := &tenablesc.AuditFile{
		BaseInfo: tenablesc.BaseInfo{
			ID:          tenablesc.ProbablyString(d.Id()),
			Name:        name,
			Description: description,
		},
	}

	afInput.Filename = filename
	afInput.OriginalFilename = name
	afInput.Version = ""

	return afInput
}
