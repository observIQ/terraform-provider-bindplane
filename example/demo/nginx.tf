resource "bindplane_source" "nginx" {
  rollout = true
  name    = "nginx-tf"
  type    = "nginx"
}

resource "bindplane_processor" "generic_delete_empty_values" {
  rollout = false
  name    = "generic-delete-empty-values-tf"
  type    = "delete_empty_values"
  parameters_json = jsonencode(
    [
      {
        "name": "telemetry_types",
        "value": ["Logs", "Metrics", "Traces"]
      },
      {
        "name": "deleted_values",
        "value": ["Null Values", "Empty Lists", "Empty Maps"]
      },
      {
        "name": "exclude_resource_keys",
        "value": []
      },
      {
        "name": "exclude_attribute_keys",
        "value": []
      },
      {
        "name": "exclude_body_keys",
        "value": []
      },
      {
        "name": "empty_string_values",
        "value": ["", "-", " ", "\t", "|4+"]
      }
    ]
  )
}

resource "bindplane_processor" "nginx_deduplicate" {
  rollout = false
  name    = "nginx-deduplicate-tf"
  type    = "log_dedup"
  parameters_json = jsonencode(
    [
      {
        "name": "interval",
        "value": 1
      },
      {
        "name": "log_count_attribute",
        "value": "log_count"
      },
      {
        "name": "exclude_fields",
        "value": [
          "body.time_local"
        ]
      },
      {
        "name": "timezone",
        "value": "UTC"
      }
    ]
  )
}

resource "bindplane_configuration" "nginx" {
  // When removing a component from a configuration and deleting that
  // component during the same apply, we want to update the configuration
  // before the component is deleted.
  lifecycle {
    create_before_destroy = true
  }

  rollout  = true
  name     = "nginx-tf"
  platform = "linux"

  source {
    name = bindplane_source.nginx.name
    processors = [
      bindplane_processor.generic_delete_empty_values.name,
      bindplane_processor.nginx_deduplicate.name
    ]
  }

  destination {
    name = bindplane_destination.gateway-east1.name
  }
}
