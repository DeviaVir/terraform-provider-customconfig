# For syntax, for forwarding configs - this allows you to make it dynamic
data "customconfig_google_forwarding_config" "your-resource" {
  ipv4_address = ["${google_compute_address.your-resource.*.address}"]
}

# this will return something `forwarding_config` will love!

# use it in your forwarding_config
resource "google_dns_managed_zone" "your-backend-service" {
  // [...]
  forwarding_config = {
    target_name_servers = ["${data.customconfig_google_forwarding_config.your-resource.target_name_servers}"]
  }
}
