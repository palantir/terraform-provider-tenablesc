resource "tenablesc_asset" "static" {
  name        = "Crash When Scanned"
  description = "Asset to use to avoid scanning devices with easily hurt feelings."
  # Now, the best way to use this asset is to generate dynamic assets that exclude this asset and scan with those.
  # Unfortunately we haven't taught the TF provider how to handle generating those, so that would be a
  # manual creation as of now.

  type = "static"

  values = [
    "10.1.1.1",
    "192.168.1.1",
  ]
}