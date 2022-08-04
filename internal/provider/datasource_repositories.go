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

func DataSourceRepositories() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRepositoriesRead,
		Description: descriptionDataSourceRepositories,
		Schema: map[string]*schema.Schema{
			"repositories": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: fmt.Sprintf(descriptionMapIDToNameTemplate, "repository", "repository"),
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"name_filter": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     ".*",
				Description: fmt.Sprintf(descriptionRegexpNameFilterTemplate, "repository"),
			},
		},
	}
}

func dataSourceRepositoriesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sc := m.(*tenablesc.Client)

	Logf(logDebug, "looking up all repositories")

	repos, err := sc.GetAllRepositories()
	if err != nil {
		return diag.FromErr(err)
	}

	Logf(logDebug, "response: %+v", repos)

	repositories := make(map[string]interface{})

	nameFilter := d.Get("name_filter").(string)

	d.SetId(fmt.Sprintf("repositories:%s", nameFilter))

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

	for _, repo := range repos {
		if nameRE.MatchString(repo.Name) {
			repositories[string(repo.ID)] = repo.Name
		}
	}

	Logf(logDebug, "Result set: %v", repositories)

	if len(repos) == 0 {
		return diag.Errorf("no repositories matching filter '^%s$'", nameFilter)
	}

	d.Set("repositories", repositories)

	return nil
}
