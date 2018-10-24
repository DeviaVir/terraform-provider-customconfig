# customconfig

Hacks for terraform.

## List

### backend services

`google_compute_region_backend_service`'s `backend` does not accept a list of
backends, since it wants a `backend { group = '.instance_group' }`. See the
examples for how to use this.

## Installation

1. Download the latest compiled binary from [GitHub releases](https://github.com/DeviaVir/terraform-provider-customconfig/releases).

1. Unzip/untar the archive.

1. Move it into `$HOME/.terraform.d/plugins`:

    ```sh
    $ mkdir -p $HOME/.terraform.d/plugins
    $ mv terraform-provider-customconfig $HOME/.terraform.d/plugins/terraform-provider-customconfig
    ```

1. Create your Terraform configurations as normal, and run `terraform init`:

    ```sh
    $ terraform init
    ```

    This will find the plugin locally.

## Development

1. `cd` into `$HOME/.terraform.d/plugins/terraform-provider-customconfig`

1. Run `dep ensure` to fetch the go vendor files

1. Make your changes

1. Run `make dev` and in your `terraform` directory, remove the current `.terraform` and re-run `terraform init`

1. Next time you run `terraform plan` it'll use your updated version
