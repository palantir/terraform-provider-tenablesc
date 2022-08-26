
data "tenablesc_plugin" "certificate_wrong_hostname" {
  # Plugin names for vulnerabilities are fixed, but this creates legibility.
  # That said, this is actually more useful for compliance plugins; auditfile
  # based plugin ids change with every upload of the audit file.
  name = "SSL Certificate with Wrong Hostname"
  # https://www.tenable.com/plugins/nessus/45411
}

resource "tenablesc_accept_risk" "certificate_wrong_hostname" {
  # ...
  plugin_id = data.tenablesc_plugin.certificate_wrong_hostname.id
  # ...
}