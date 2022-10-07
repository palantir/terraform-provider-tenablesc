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

const (
	descriptionDataSourceNameFindTemplate = `Name of the %s to find.`

	descriptionRegexpNameFilterTemplate = `A regexp-based filter to match target %s names. 
					 Will be wrapped in ^ and $ before compilation. 
					 If not given, will return all elements.`

	descriptionMapIDToNameTemplate = `A map of %s IDs to %s names`
)
const (
	descriptionDefaultDescriptionValue = "Managed by Terraform"

	// For top-level descriptions, use complete sentences including periods(!) to describe duty.
	// Include also which credential scope is required if relevant.
	// This may be skipped if both admin and org credentials are usable.
	descriptionOrgCredentialsRequired = `
Requires Organization credentials.`
	descriptionAdminCredentialsRequired = `
Requires Administrator (org=0) credentials.`

	// Data Sources

	descriptionDataSourceAsset              = `Look up an asset by name field.` + descriptionOrgCredentialsRequired
	descriptionDataSourceAssets             = `Look up a set of asset IDs based on a regular expression name filter.` + descriptionOrgCredentialsRequired
	descriptionDataSourceCredential         = `Look up a credential object ID by name field.`
	descriptionDataSourcePlugin             = `Look up a plugin ID based on name.`
	descriptionDataSourceRepositories       = `Look up a set of repositories based on a regular expression name filter.`
	descriptionDataSourceRepository         = `Look up a repository ID based on name.`
	descriptionDataSourceScanPolicyTemplate = `Look up a scan policy template ID based on name.`

	// Resources
	descriptionResourceAcceptRisk                        = `Create and manage Accept Risk Rules.` + descriptionOrgCredentialsRequired
	descriptionResourceAsset                             = `Create and manage Assets.` + descriptionOrgCredentialsRequired
	descriptionResourceAuditFile                         = `Create and manage Audit Files.`
	descriptionResourceOrganization                      = `Create and manage Organizations.` + descriptionAdminCredentialsRequired
	descriptionResourceOrganizationScanZoneAssociation   = `Manage Scan Zones associated to an Organization.` + descriptionAdminCredentialsRequired
	descriptionResourceRecastRisk                        = `Create and manage Recast Risk Rules.` + descriptionOrgCredentialsRequired
	descriptionResourceRepository                        = `Create and Manage Repositories.` + descriptionAdminCredentialsRequired
	descriptionResourceRepositoryOrganizationAssociation = `Manage Organization access to Repositories.` + descriptionAdminCredentialsRequired
	descriptionResourceRole                              = `Create and Manage User Roles.` + descriptionOrgCredentialsRequired
	descriptionResourceScan                              = `Create and Manage Scans.` + descriptionOrgCredentialsRequired
	descriptionResourceScanPolicy                        = `Create and Manage Scan Policies.` + descriptionOrgCredentialsRequired
	descriptionResourceScanZone                          = `Create and Manage Scan Zones.` + descriptionAdminCredentialsRequired

	// Fields
	// Field descriptions should be brief and self-descriptive phrases, even if slightly redundant.
	// Complex fields should include schema descriptions here.

	// Field Names and Descriptions
	descriptionAssetName               = `Asset name`
	descriptionAssetDescription        = `Asset description`
	descriptionAuditFileName           = `Audit file Name as presented in SC`
	descriptionAuditFileDescription    = `Audit File description`
	descriptionPluginName              = `Plugin name`
	descriptionRepositoryName          = `Repository name`
	descriptionRepositoryDescription   = `Repository description`
	descriptionOrganizationName        = `Organization name`
	descriptionOrganizationDescription = `Organization description`
	descriptionRoleName                = `Role name`
	descriptionRoleDescription         = `Role description`
	descriptionScanPolicyName          = `Scan Policy name`
	descriptionScanPolicyDescription   = `Scan Policy description`
	descriptionScanZoneName            = `Scan Zone name`
	descriptionScanZoneDescription     = `Scan Zone description`

	// ID fields
	descriptionRepositoryID            = "Repository ID"
	descriptionPluginID                = "Plugin ID"
	descriptionOrganizationID          = `Organization ID`
	descriptionScanPolicyTemplateID    = `Scan Policy Template ID`
	descriptionAuditFileID             = `Audit File ID`
	descriptionOrganizationScanZoneIDs = `Scan Zone IDs to be allowed to be used by organization`

	// Miscellaneous
	descriptionAssetDefinedIPs      = `IP addresses defined in the asset`
	descriptionAssetDefinedDNSNames = `DNS Names defined in the asset`
	descriptionAssetType            = `Asset type - may be 'dnsname' or 'static'`
	descriptionAssetValues          = `Asset values - must be either DNS names or IPs based on type of asset.`

	descriptionAuditFileContent    = `Audit file content`
	descriptionAuditFileSCFilename = `Filename of audit file as stored in SC`

	descriptionRiskRuleHostType  = `Host Type may be 'all', 'ip', or 'asset'`
	descriptionRiskRuleHostValue = `A list of values depending on the host type.
  * Must be empty for type 'all'; 
  * For 'ip' must be a list of IP addresses
  * For 'asset' must be a list of asset IDs.`
	descriptionPort     = `Network port`
	descriptionProtocol = `Network protocol. Default: 'any' 
  * tcp
  * udp
  * icmp
  * unknown 
  * any `
	descriptionComments             = `Comments`
	descriptionAcceptRiskExpiration = `Expiration date for accept risk rule in RFC3339 format`

	descriptionOrganizationZoneSelection = `Scan Zone Selection for organization. May be:
 * auto_only
 * locked
 * selectable
 * selectable+auto
 * selectable+auto_restricted `
	descriptionOrganizationRestrictedIPs = `If provided, limits IPs allowed in zone to list. May be provided as IPs, CIDRs, or ranges.`

	descriptionRecastNewSeverity = `Updated severity for ticket in numeric form. 
  * 0 - Info
  * 1 - Low
  * 2 - Medium
  * 3 - High
  * 4 - Critical`

	descriptionRepositoryIPRange     = `Range of IPs allowed to be stored in the repository - may be CIDR or Range format`
	descriptionTrendingDays          = `Days to store trend data`
	descriptionTrendWithRaw          = `Store raw data with trends`
	descriptionVulnerabilityLifetime = `Specify custom storage durations in days for types of vulnerabilities`

	descriptionGroupAssignment = `Access within organization to grant to repository. Valid values are:
 * all
 * fullAccess
 * partial`

	descriptionScanPolicyPreferences = `Key-value map of preferences to set and their values. Refer to documentation and browser developer tools to get preference names`
	descriptionScanPolicyFamilies    = `Plugin Families to include in scan`
	descriptionScanPolicyTag         = `Tag for scan policy`

	descriptionScanZoneCIDRs = `CIDR blocks included in scan zone`
)
