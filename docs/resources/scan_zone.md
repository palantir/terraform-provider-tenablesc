---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tenablesc_scan_zone Resource - terraform-provider-tenablesc"
subcategory: ""
description: |-
  Create and Manage Scan Zones.
  Requires Administrator (org=0) credentials.
---

# tenablesc_scan_zone (Resource)

Create and Manage Scan Zones.
Requires Administrator (org=0) credentials.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Scan Zone name
- `zone_cidrs` (Set of String) CIDR blocks included in scan zone

### Optional

- `description` (String) Scan Zone description

### Read-Only

- `id` (String) The ID of this resource.

