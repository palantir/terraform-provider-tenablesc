data "tenablesc_scan_policy_template" "advanced" {
  name = "Advanced Scan"
  # This is a built-in scan policy base for freeform vuln scans.
}

resource "tenablesc_scan_policy" "advanced_vulnerability_scan" {
  # ...
  policy_template_id = data.tenablesc_scan_policy_template.advanced.id
  # ...
}