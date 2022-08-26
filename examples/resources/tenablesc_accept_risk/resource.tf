
locals {
  endpoint_asset_ids = [for id, name in data.tenablesc_assets.endpoints : id]
  repository_ids     = [for id, name in data.tenablesc_repositories.endpoints : id]
}

data "tenablesc_assets" "endpoints" {
  # assuming you've generated assets of some sort to identify desktops you're scanning.
  name_filter = ".*-endpoints"
}

data "tenablesc_repositories" "endpoints" {
  # also assuming you're scanning them into individual repositories, probably
  # to make it easier to grant access to the proper endpoint management teams.
  name_filter = ".*-endpoints"
}

resource "tenablesc_accept_risk" "self_signed_certificate" {
  # let's also assume, notionally, we're not interested in whatever weird certs
  # users might generate for themselves being signed properly.
  for_each   = toset(local.endpoint_asset_ids)
  host_type  = "asset"
  host_value = each.value

  repository_ids = local.repository_ids
  # https://www.tenable.com/plugins/nessus/45411
  plugin_id = "45411"
}