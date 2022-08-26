

data "tenablesc_credential" "ssh" {
  name = "ssh user credential"
}

resource "tenablesc_scan" "ssh_scan" {
  # ...

  credential_ids = [data.tenablesc_credential.ssh.id]

  # ...
}