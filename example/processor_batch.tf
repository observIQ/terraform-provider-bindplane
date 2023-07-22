resource "bindplane_processor" "batch-default" {
  rollout = false
  name = "example-batch"
  type = "batch"
}

resource "bindplane_processor" "batch-custom" {
  rollout = false
  name = "example-batch-custom"
  type = "batch"
  parameters_json = jsonencode(
    [
      {
        "name": "send_batch_size",
        "value": 200
      },
      {
        "name": "send_batch_max_size",
        "value": 400
      },
      {
        "name": "timeout",
        "value": "2s"
      }
    ]
  )
}

