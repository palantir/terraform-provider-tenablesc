resource "tenablesc_role" "scan_only" {
  name        = "Scan Only"
  description = "Role-holder may only run scans."

  scan        = true
  agents_scan = true
}