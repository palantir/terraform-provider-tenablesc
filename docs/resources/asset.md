---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tenablesc_asset Resource - terraform-provider-tenablesc"
subcategory: ""
description: |-
  Create and manage Assets.
  Requires Organization credentials.
---

# tenablesc_asset (Resource)

Create and manage Assets.
Requires Organization credentials.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Asset name
- `type` (String) Asset type - may be 'dnsname' or 'static'

### Optional

- `description` (String) Asset description
- `values` (Set of String) Asset values - must be either DNS names or IPs based on type of asset.

### Read-Only

- `id` (String) The ID of this resource.

