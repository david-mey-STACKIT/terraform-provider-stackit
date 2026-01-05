variable "network_name_listener" {
  description = "The STACKIT Network name"
  type        = string
  default     = "network_max_listener"
}

variable "network_name_target" {
  description = "The STACKIT Network name"
  type        = string
  default     = "network_max_target"
}

variable "network_role_listeners" {
  description = "The default network role"
  type        = string
  default     = "ROLE_LISTENERS"
}

variable "network_role_targets" {
  description = "The default network role"
  type        = string
  default     = "ROLE_TARGETS"
}

variable "server_name_max" {
  description = "The name of the backend server"
  type        = string
  default     = "backend_server_max"
}

variable "disable_security_group_assignment" {
  description = "disable target security group assignment"
  type        = bool
  default     = true
}

variable "protocol_https" {
  default = "PROTOCOL_HTTPS"
}

variable "private_network_only" {}
variable "acl" {}
variable "ephemeral_address" {}

variable "observability_logs_push_url" {}
variable "observability_metrics_push_url" {}
variable "observability_credential_name" {}
variable "observability_credential_username" {}
variable "observability_credential_password" {}

resource "stackit_network" "listener_network" {
  project_id       = var.project_id
  name             = var.network_name_listener
  ipv4_nameservers = ["8.8.8.8"]
  ipv4_prefix      = "10.11.10.0/24"
  routed           = "true"
}

resource "stackit_network" "target_network" {
  project_id       = var.project_id
  name             = var.network_name_target
  ipv4_nameservers = ["8.8.8.8"]
  ipv4_prefix      = "10.11.1.0/24"
  routed           = "true"
}

resource "stackit_network_interface" "network_interface_listener" {
  project_id = var.project_id
  network_id = stackit_network.listener_network.network_id
  lifecycle {
    ignore_changes = [
      security_group_ids,
    ]
  }
}

resource "stackit_network_interface" "network_interface_target" {
  project_id = var.project_id
  network_id = stackit_network.target_network.network_id
  lifecycle {
    ignore_changes = [
      security_group_ids,
    ]
  }
}

resource "stackit_public_ip" "public_ip" {
  project_id           = var.project_id
  network_interface_id = stackit_network_interface.network_interface_listener.network_interface_id
  lifecycle {
    ignore_changes = [
      network_interface_id
    ]
  }
}

resource "stackit_server" "server_max" {
  project_id        = var.project_id
  availability_zone = var.availability_zone
  name              = var.server_name_max
  machine_type      = var.machine_type
  boot_volume = {
    size                  = 20
    source_type           = "image"
    source_id             = var.image_id
    delete_on_termination = "true"
  }
  network_interfaces = [
    stackit_network_interface.network_interface_target.network_interface_id
  ]
  # Explicit dependencies to ensure ordering
  depends_on = [
    stackit_network.target_network,
    stackit_network_interface.network_interface_target
  ]
}

resource "stackit_loadbalancer_observability_credential" "observer" {
  project_id   = var.project_id
  display_name = var.observability_credential_name
  password     = var.observability_credential_password
  username     = var.observability_credential_username
}

resource "stackit_alb" "loadbalancer" {
  region                            = var.region
  project_id                        = var.project_id
  name                              = var.loadbalancer_name
  plan_id                           = var.plan_id
  disable_security_group_assignment = var.disable_security_group_assignment
  target_pools = [
    {
      name = var.target_pool_name
      active_health_check = {
        interval            = "1s"
        interval_jitter     = "9.990s"
        timeout             = "0.020s"
        healthy_threshold   = 1
        unhealthy_threshold = 1
        http_health_checks = {
          ok_status = ["200", "201"]
          path      = "/lol"
        }
      }
      target_port = var.target_port
      targets = [
        {
          display_name = var.target_display_name
          ip           = stackit_network_interface.network_interface.ipv4
        }
      ]
      tls_config = {
        enabled                     = true
        skip_certificate_validation = true
      }
    }
  ]
  listeners = [{
    port = var.listener_port
    http = {
      hosts = [{
        host = "*"
        rules = [{
          target_pool = var.target_pool_name
          query_parameters = [{
            name        = "a_query_parameter"
            exact_match = "value"
            }, {
            name        = "one-more"
            exact_match = "query_parameters_1337"
          }]
          headers = [{
            name        = "a-header"
            exact_match = "value"
            }, {
            name        = "one-more2"
            exact_match = "header"
            }, {
            name = "one-more1"
          }]
          path = {
            prefix = "/"
          }
          cookie_persistence = {
            name = "a-name"
            ttl  = "1s"
          }
          }, {
          path = {
            prefix = "/"
          }
          cookie_persistence = {
            name = "a-name"
            ttl  = "10000000s"
          }
          target_pool = "my-pool"
        }]
      }]
    }
    https = {}
    protocol = var.protocol_https
  }]
  networks = [
    {
      network_id = stackit_network.listener_network.network_id
      role       = var.network_role_listeners
    },
    {
      network_id = stackit_network.target_network.network_id
      role       = var.network_role_targets
    }
  ]
  options = {
    ephemeral_address = var.ephemeral_address
    private_network_only = var.private_network_only
    acl                  = [var.acl]
    observability = {
      logs = {
        credentials_ref = stackit_loadbalancer_observability_credential.observer.credentials_ref
        push_url        = var.observability_logs_push_url
      }
      metrics = {
        credentials_ref = stackit_loadbalancer_observability_credential.observer.credentials_ref
        push_url        = var.observability_metrics_push_url
      }
    }
  }
}
