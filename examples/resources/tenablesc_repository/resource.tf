resource "tenablesc_repository" "lab" {
  name     = "Lab"
  ip_range = "0.0.0.0/0"
}

data "tenablesc_organization" "lab" {
  name = "Lab"
}

resource "tenablesc_repository_organization_association" "lab" {
  repository_id = tenablesc_repository.lab.id
  organization = [
    {
      organization_id = data.tenablesc_organization.lab.id
      # group_assignment = "all" || "fullAccess" || "partial"
    }
  ]
}