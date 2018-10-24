# customconfig

Hacks for terraform.

## List

### backend services

`google_compute_region_backend_service`'s `backend` does not accept a list of
backends, since it wants a `backend { group = '.instance_group' }`. See the
examples for how to use this.
