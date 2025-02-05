locals {
  cluster_name = "ecom-east1"
}

resource "bindplane_source" "k8s_container" {
  rollout = false
  name    = "kubernetes-container-tf"
  type    = "k8s_container"
  parameters_json = jsonencode(
    [
      {
        "name": "cluster_name"
        "value": "${local.cluster_name}"
      },
      {
        "name": "exclude_file_path"
        "value": [
          "/var/log/containers/observiq-*-collector-*",
          "/var/log/containers/bindplane-*-agent-*",
          "/var/log/containers/*kube-system*",
        ]
      }
    ]
  )
}

resource "bindplane_source" "k8s_kubelet" {
  rollout = false
  name    = "kubernetes-kubelet-tf"
  type    = "k8s_kubelet"
  parameters_json = jsonencode(
    [
      {
        "name": "cluster_name"
        "value": "${local.cluster_name}"
      }
    ]
  )
}

resource "bindplane_source" "k8s_prometheus" {
  rollout = false
  name = "kubernetes-prometheus-tf"
  type = "k8s_prometheus_node"
  parameters_json = jsonencode(
    [
      {
        "name": "cluster_name"
        "value": "${local.cluster_name}"
      }
    ]
  )
}

resource "bindplane_source" "k8s_events" {
  rollout = false
  name    = "kubernetes-events-tf"
  type    = "k8s_events"
  parameters_json = jsonencode(
    [
      {
        "name": "cluster_name"
        "value": "${local.cluster_name}"
      }
    ]
  )
}

resource "bindplane_source" "k8s_cluster" {
  rollout = false
  name    = "kubernetes-cluster-tf"
  type    = "k8s_cluster"
  parameters_json = jsonencode(
    [
      {
        "name": "cluster_name"
        "value": "${local.cluster_name}"
      }
    ]
  )
}

resource "bindplane_source" "k8s_gateway" {
  rollout = false
  name    = "kubernetes-gateway-tf"
  type    = "bindplane_gateway"
}

resource "bindplane_source" "k8s_otlp_gateway" {
  rollout = false
  name    = "kubernetes-otlp-tf"
  type    = "otlp"
}

resource "bindplane_destination" "k8s_gateway" {
  rollout = false
  name    = "kubernetes-gateway-tf"
  type    = "bindplane_gateway"
  parameters_json = jsonencode(
    [
      {
        "name" : "hostname",
        "value" : "bindplane-gateway-agent.bindplane-agent.svc.cluster.local"
      },
      {
        "name" : "retry_on_failure_enabled",
        "value" : false
      },
      {
        "name" : "sending_queue_enabled",
        "value" : false
      },
      {
        "name" : "persistent_queue_enabled",
        "value" : false
      },
    ]
  )
}

resource "bindplane_configuration" "k8s-node" {
  lifecycle {
    create_before_destroy = true
  }

  rollout  = true
  name     = "k8s-node-tf"
  platform = "kubernetes-daemonset"

  source {
    name = bindplane_source.k8s_container.name
  }

  source {
    name = bindplane_source.k8s_kubelet.name
  }

  source {
    name = bindplane_source.k8s_prometheus.name
  }

  source {
    name = bindplane_source.k8s_otlp_gateway.name
  }

  destination {
    name = bindplane_destination.k8s_gateway.name
  }
}

resource "bindplane_configuration" "k8s-cluster" {
  lifecycle {
    create_before_destroy = true
  }

  rollout  = true
  name     = "k8s-cluster-tf"
  platform = "kubernetes-deployment"

  source {
    name = bindplane_source.k8s_cluster.name
  }

  source {
    name = bindplane_source.k8s_events.name
  }

  destination {
    name = bindplane_destination.k8s_gateway.name
  }
}

resource "bindplane_configuration" "k8s-gateway" {
  lifecycle {
    create_before_destroy = true
  }

  rollout  = true
  name     = "k8s-gateway-tf"
  platform = "kubernetes-gateway"

  source {
    name = bindplane_source.k8s_gateway.name
  }

  destination {
    name = bindplane_destination.k8s_gateway.name
  }
}
