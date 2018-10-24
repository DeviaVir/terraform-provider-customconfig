# For syntax, for backend services - this allows you to make it dynamic
data "customconfig_google_backend" "your-resource" {
  instance_groups = ["${google_compute_instance_group_manager.your-resource.*.instance_group}"]
}

# this will return something `backend` will love!

# use it in your backend service
resource "google_compute_region_backend_service" "your-backend-service" {
  // [...]
  backend = ["${data.customconfig_google_backend.your-service.backends}"]
}
