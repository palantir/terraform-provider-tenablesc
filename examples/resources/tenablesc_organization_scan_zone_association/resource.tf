resource "tenablesc_scan_zone" "lab_cidrs" {
  name        = "Lab"
  description = "Lab-only cidrs for lab org scans"
  zone_cidrs = [
    "192.168.1.0/24",
  ]
}

resource "tenablesc_organization" "lab" {
  name = "Labs"

}

resource "tenablesc_organization_scan_zone_association" "lab" {
  organization_id = tenablesc_organization.lab
  scan_zone_ids   = [tenablesc_scan_zone.lab_cidrs.id]
}

# You may notice we don't have a tenablesc_scanner resource. Because we don't want to put the scanner passwords in
# TF state, we build scanners separately. The missing step in this example is associating the desired scanners
# to the 'Lab' scan zone.