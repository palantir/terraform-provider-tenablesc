terraform {
  required_providers {
    tenablesc = {
      source = "palantir/tenablesc"
    }
  }
}

provider "tenablesc" {
  uri        = "https://your_sc_host.dns.name/rest" # may be specified with TENABLESC_URI environment variable
  access_key = ""                                   # may be specified with TENABLESC_ACCESS_KEY environment variable
  secret_key = ""                                   # may be specified with TENABLESC_SECRET_KEY environment variable
}

data "tenablesc_repository" "default" {
  name = "default"
}

output "tenablesc_default_repository_id" {
  value = tenablesc_repository.default.id
}
