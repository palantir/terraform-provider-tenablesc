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

func DataSourceRepository() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRepositoryRead,
		Description: descriptionDataSourceRepository,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf(descriptionDataSourceNameFindTemplate, "repository"),
			},
		},
	}
}

func dataSourceRepositoryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	sc := m.(*tenablesc.Client)

	repoName := d.Get("name").(string)

	Logf(logDebug, "looking up %s", repoName)

	repos, err := sc.GetAllRepositories()
	if err != nil {
		return diag.FromErr(err)
	}

	Logf(logDebug, "response: %+v", repos)

	for _, repo := range repos {
		Logf(logTrace, "comparing repository: %+v", *repo)
		if strings.Compare(repo.Name, repoName) == 0 {
			d.SetId(string(repo.ID))
			return nil
		}
	}

	return diag.Errorf("No repository found with name like [%s]", repoName)
}
