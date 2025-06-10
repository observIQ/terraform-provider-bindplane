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
  password = "password"
}

resource "bindplane_source" "host" {
  rollout = true
  name = "my-host"
  type = "host"
  parameters_json = jsonencode(
    [
      {
        "name": "collection_interval",
        "value": 32
      },
      {
        "name": "enable_process",
        "value": false
      }
    ]
  )
}

resource "bindplane_source" "journald" {
  rollout = true
  name = "my-journald"
  type = "journald"
}

resource "bindplane_destination" "custom" {
  rollout = true
  name = "example-custom"
  type = "custom"
  parameters_json = jsonencode(
    [
      {
        "name": "telemetry_types",
        "value": [
          "Logs",
        ]
      },
      {
        "name": "configuration",
        "value": <<EOT
logging:
  verbosity: detailed
  sampling_initial: 5
  sampling_thereafter: 200
EOT
      }
    ]
  )
}

resource "bindplane_processor" "batch" {
  rollout = true
  name = "my-batch"
  type = "batch"
}

resource "bindplane_configuration" "configuration" {
  lifecycle {
    create_before_destroy = true
  }

  advanced {
    metrics {
      port = 8885
      level = "normal"
    }
  }

  measurement_interval = "1m"

  rollout = true

  rollout_options {
    type = "progressive"
    parameters {
      name = "stages"
      value {
        labels = {
          env = "stage"
        }
        name = "stage"
      }
      value {
        labels = {
          env = "production"
        }
        name = "production"
      }
    }
  }

  name = "my-config"
  platform = "linux"
  labels = {
    environment = "production"
    managed-by  = "terraform"
  }

  source {
    name = bindplane_source.journald.name
    processors = [
      bindplane_processor_bundle.bundle.name,
    ]
  }

  source {
    name = bindplane_source.host.name
  }

  destination {
    name = bindplane_destination.custom.name
    processors = [
      bindplane_processor.batch.name,

      // order matters here
      bindplane_processor.time-parse-http-datatime.name
    ]
  }

  extensions = [
    bindplane_extension.pprof.name
  ]
}

resource "bindplane_processor" "json-parse-body" {
  rollout = false
  name = "json-parse-body"
  type = "parse_json"
  parameters_json = jsonencode(
    [
      {
        "name": "telemetry_types",
        "value": [
          "Logs",
        ]
      },
      {
        "name": "log_source_field_type",
        "value": "Body"
      },
      {
        "name": "log_body_source_field",
        "value": ""
      },
      {
        "name": "log_target_field_type",
        "value": "Body"
      }
    ]
  )
}

resource "bindplane_processor" "time-parse-http-datatime" {
  rollout = false
  name = "time-parse-http-datatime"
  type = "parse_timestamp"
  parameters_json = jsonencode(
    [
      {
        "name": "telemetry_types",
        "value": [
          "Logs",
        ]
      },
      {
        "name": "log_field_type",
        "value": "Body"
      },
      {
        "name": "log_source_field",
        "value": "datetime"
      },
      {
        "name": "log_time_format",
        "value": "Manual"
      },
      {
        "name": "log_epoch_layout",
        "value": "s"
      },
      {
        "name": "log_manual_timestamp_layout",
        "value": "%d/%b/%Y:%H:%M:%S %z"
      }
    ]
  )
}

resource "bindplane_extension" "pprof" {
  rollout = true
  name = "my-pprof"
  type = "pprof"
  parameters_json = jsonencode(
    [
      {
        "name": "listen_address",
        "value": "0.0.0.0"
      },
      {
        "name": "tcp_port",
        "value": 5000,
      },
    ]
  )
}

resource "bindplane_processor_bundle" "bundle" {
  rollout = true
  name = "my-bundle"

  processor {
    name = bindplane_processor.batch.name
  }

  processor {
    name = bindplane_processor.time-parse-http-datatime.name
  }
}

resource "bindplane_connector" "routing" {
  rollout = true
  name = "log-router"
  type = "routing"
  parameters_json = jsonencode(
    [
      {
        "name": "telemetry_types",
        "value": [
          "Logs"
        ]
      },
      {
        "name": "routes",
        "value": [
          {
            "condition": {
              "ottl": "(attributes[\"env\"] == \"prod\")",
              "ottlContext": "resource",
              "ui": {
                "operator": "",
                "statements": [
                  {
                    "key": "env",
                    "match": "resource",
                    "operator": "Equals",
                    "value": "prod"
                  }
                ]
              }
            },
            "id": "datadog"
          },
          {
            "condition": {
              "ottl": "(attributes[\"env\"] == \"dev\")",
              "ottlContext": "resource",
              "ui": {
                "operator": "",
                "statements": [
                  {
                    "key": "env",
                    "match": "resource",
                    "operator": "Equals",
                    "value": "dev"
                  }
                ]
              }
            },
            "id": "google"
          },
          {
            "condition": {
              "ottl": "",
              "ui": {
                "operator": "",
                "statements": [
                  {
                    "key": "",
                    "match": "attributes",
                    "operator": "Equals",
                    "value": ""
                  }
                ]
              }
            },
            "id": "fallback"
          }
        ]
      }
    ] 
  )
}

resource "bindplane_connector" "fluent_router" {
  rollout = true
  name = "fluent-router"
  type = "routing"
  parameters_json = jsonencode(
    [
      {
        "name": "telemetry_types",
        "value": [
          "Logs"
        ]
      },
      {
        "name": "routes",
        "value": [
          {
            "condition": {
              "ottl": "(attributes[\"env\"] == \"prod\")",
              "ottlContext": "resource",
              "ui": {
                "operator": "",
                "statements": [
                  {
                    "key": "env",
                    "match": "resource",
                    "operator": "Equals",
                    "value": "prod"
                  }
                ]
              }
            },
            "id": "parser"
          },
          {
            "condition": {
              "ottl": "(attributes[\"env\"] == \"dev\")",
              "ottlContext": "resource",
              "ui": {
                "operator": "",
                "statements": [
                  {
                    "key": "env",
                    "match": "resource",
                    "operator": "Equals",
                    "value": "dev"
                  }
                ]
              }
            },
            "id": "loki"
          },
        ]
      }
    ] 
  )
}

resource "bindplane_configuration_v2" "configuration" {
  lifecycle {
    create_before_destroy = true
  }

  rollout = true

  name = "my-config-v2"
  platform = "linux"

  source {
    name = bindplane_source.otlp.name

    route {
      route_id = "metric-batcher"
      telemetry_type = "metrics"
      components = [
        "processors/batcher"
      ]
    }

    route {
      route_id = "log-parser"
      telemetry_type = "logs"
      components = [
        "processors/parser"
      ]
    }

    route {
      route_id = "trace-batcher"
      telemetry_type = "traces"
      components = [
        "processors/batcher"
      ]
    }
  }

  source {
    name = bindplane_source.journald.name
    processors = [
      bindplane_processor_bundle.bundle.name,
    ]

    route {
      route_id = "log-parser"
      telemetry_type = "logs"
      components = [
        "processors/parser"
      ]
    }

    route {
      route_id = "metric-batcher"
      telemetry_type = "metrics"
      components = [
        "processors/batcher"
      ]
    }

    route {
      route_id = "trace-batcher"
      telemetry_type = "traces"
      components = [
        "processors/batcher"
      ]
    }
  }

  source {
    name = bindplane_source.fluent.name
    route {
      route_id = "fluent-router"
      telemetry_type = "logs"
      components = [
        "connectors/fluent-router"
      ]
    }
  }

  source {
    name = bindplane_source.host.name
    route {
      route_id = "batch-metrics"
      telemetry_type = "metrics"
      components = [
        "processors/batcher"
      ]
    }
  }

  processor_group {
    route_id = "parser"
    processors = [
      bindplane_processor.json-parse-body.name,
      bindplane_processor.time-parse-http-datatime.name
    ]
    route {
      route_id = "log-batcher"
      telemetry_type = "logs"
      components = [
        "processors/batcher"
      ]
    }
  }

  processor_group {
    route_id = "batcher"
    processors = [
      bindplane_processor.batch.name
    ]
    route {
      route_id = "log-destinations"
      telemetry_type = "logs"
      components = [
        "connectors/logging-router"
      ]
    }
    route {
      route_id = "metric-destinations"
      telemetry_type = "metrics"
      components = [
        "destinations/${bindplane_destination.datadog.id}",
        "destinations/${bindplane_destination.google.id}",
        "destinations/${bindplane_destination.loki.id}"
      ]
    }
    route {
      route_id = "trace-destinations"
      telemetry_type = "traces"
      components = [
        "destinations/${bindplane_destination.datadog.id}",
        "destinations/${bindplane_destination.google.id}",
        "destinations/${bindplane_destination.loki.id}"
      ]
    }
  }

  connector {
    route_id = "logging-router"
    name = bindplane_connector.routing.name
    route {
      route_id = "datadog"
      telemetry_type = "logs"
      components = [
        "destinations/${bindplane_destination.datadog.id}",
      ]
    }
    route {
      route_id = "google"
      telemetry_type = "logs"
      components = [
        "destinations/${bindplane_destination.google.id}",
      ]
    }
    route {
      route_id = "fallback"
      telemetry_type = "logs"
      components = [
        "destinations/${bindplane_destination.loki.id}"
      ]
    }
  }

  connector {
    route_id = "fluent-router"
    name = bindplane_connector.fluent_router.name
    route {
      route_id = "parser"
      telemetry_type = "logs"
      components = [
        "processors/parser"
      ]
    }
    route {
      route_id = "loki"
      telemetry_type = "logs"
      components = [
        "destinations/${bindplane_destination.loki.id}"
      ]
    }
  }

  destination {
    route_id   = bindplane_destination.google.id
    name = bindplane_destination.google.name
    processors = [
      bindplane_processor.batch.name,
      bindplane_processor.time-parse-http-datatime.name
    ]
  }

  destination {
    route_id = bindplane_destination.datadog.id
    name = bindplane_destination.datadog.name
  }

  destination {
    route_id   = bindplane_destination.loki.id
    name = bindplane_destination.loki.name
  }

  extensions = [
    bindplane_extension.pprof.name
  ]
}
