terraform {
  backend "gcs" {
    bucket = "terraform-provider-bindplane-state"
    prefix = "state"
  }
}
