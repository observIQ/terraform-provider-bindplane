# Bindplane Terraform Provider Local Test Folder

## Before testing:

Run:

- `export TF_CLI_CONFIG_FILE=./dev.tfrc`
- `make provider`
- in `test/local`:
  - `terraform init`

## Before applying changes:

Run:

- `make ci-local`

- in `test/local`:

  - `terraform destroy`
  - `terraform apply`

- To auto approve:

  - add `-auto-approve` flag

- You can also run `make test-local-apply` to run all of these at once.
