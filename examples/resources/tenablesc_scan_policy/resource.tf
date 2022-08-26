
locals {
  os_vulnerability_families = [
    # Not going to provide this list here as it will vary.
    # To collect, go build your initial policy, then look at it with curl or
    # browser developer tools to collect the correct IDs.
  ]

  preferences = {
    # Same as above, this is an organizational decision.
    # create your initial policy, then collect the settings from there.
    # Audit regularly in case new options have been added.
  }
}


data "sc_scan_policy_template" "advanced" {
  name = "Advanced Scan"
}

resource "sc_scan_policy" "vulnerability_scan_port22" {
  name               = "TF Vulnerability - Ubuntu, CentOS, RHEL (port 22)"
  policy_template_id = data.sc_scan_policy_template.advanced.id
  families           = local.os_vulnerability_families
  preferences        = local.preferences
}
