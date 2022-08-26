resource "tenablesc_auditfile" "stig_ubuntu_20_04" {
  # Holding auditfiles in code and uploading them instead of pulling direct from vendor has many advantages.
  # Allows you to uptake recasts at the same time as updated audit files, and have a pinned known-good old version.
  # Keeps history in sync in-repo end-to-end.

  name    = "STIG Ubuntu 20.04"
  content = file("path/in/repo/to/file.audit")
}