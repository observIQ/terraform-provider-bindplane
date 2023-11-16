terraform {
  required_providers {
    bindplane = {
      source = "observiq/bindplane"
    }
  }
}

provider "bindplane" {
  remote_url = "https://localhost:3100"
  username = "tfu"
  password = "tfp"

  // server's certificate is signed by this CA
  tls_certificate_authority = "../../client/tls/bindplane-ca.crt"

  // mtls client auth
  tls_certificate = "../../client/tls/bindplane-client.crt"
  tls_private_key = "../../client/tls/bindplane-client.key"

  // invalid mtls, client ca is not trusted by the server
  // tls_certificate = "../../client/tls/test-client.crt"
  // tls_private_key = "../../client/tls/test-client.key"
}

module "test" {
  source = "./modules/test"
}
