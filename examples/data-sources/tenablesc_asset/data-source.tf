
locals {
  severity_critical = 4
}

data "tenablesc_asset" "existing" {
  name = "existing asset name"
}

data "tenablesc_repository" "main" {
  name = "main"
}

resource "tenablesc_recast_risk" "tls_1_1_deprecated" {
  new_severity = local.severity_critical
  host_type    = "asset"
  host_value   = data.tenablesc_asset.existing.id
  # https://www.tenable.com/plugins/nessus/157288
  plugin_id = "157288"
}