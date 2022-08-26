data "tenablesc_scan_policy" "basic" {
  name = "Basic Network Scan"
}

data "tenablesc_repository" "lab" {
  name = "Lab"
}

data "tenablesc_asset" "lab" {
  name = "Lab"
}

data "tenablesc_credential" "lab" {
  name = "Lab"
}

resource "tenablesc_scan" "basic" {
  name = "Nightly Lab Basic Scan"

  repository_id = data.tenablesc_repository.lab.id
  policy_id     = data.tenablesc_scan_policy.basic.id

  asset_ids      = [data.tenablesc_asset.lab]
  credential_ids = [data.tenablesc_credential.lab.id]

  # easiest way to determine what this should look like is to
  # use developer tools in the browser, create a scan, and look at the result.
  schedule_repeat_rule = "FREQ=DAILY;INTERVAL=1"                 #Every day
  schedule_start       = "TZID=America/New_York:20190909T200000" #Start at 8pm ET, no earlier than September 9 2019.
}