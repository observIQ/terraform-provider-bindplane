terraform {
  required_providers {
    bindplane = {
      source = "observiq/bindplane"
    }
  }
}

provider "bindplane" {
  remote_url = "http://localhost:3001"
  username = "admin"
  password = "admin"
}
