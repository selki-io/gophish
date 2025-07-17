// ====================
// Locals
// ====================

locals {
  project = "${var.env}-selki"

  hostname = {
    dev  = "${var.service}-dev"
    qa   = "${var.service}-qa"
    stag = "${var.service}-stag"
    prod = "${var.service}"
  }

  ingress = {
    dev  = "INGRESS_TRAFFIC_ALL"
    qa   = "INGRESS_TRAFFIC_ALL"
    stag = "INGRESS_TRAFFIC_ALL"
    prod = "INGRESS_TRAFFIC_ALL"
  }

  es_host = {
    dev  = "https://d202584ccf854f89af3477767b69558b.es.us-east-1.aws.elastic.cloud:443"
    prod = "https://d202584ccf854f89af3477767b69558b.es.us-east-1.aws.elastic.cloud:443"
  }

  es_index = {
    dev  = "reaper-000001"
    prod = "reaper-000001"
  }

  min_count = {
    dev  = 0
    qa   = 0
    stag = 0
    prod = 1
  }

  max_count = {
    dev  = 1
    qa   = 1
    stag = 1
    prod = 10
  }

  clickup_notification_views = {
    dev  = "8chwgv6-2011"
    qa   = ""
    stag = ""
    prod = "8chwgv6-1991"
  }
}

// ====================
// Config
// ====================

provider "google" {
  project = local.project
  region  = var.location
}

// ====================
// Init
// ====================

terraform {
  backend "gcs" {
    prefix = "bleeper"
  }

  required_version = ">= 1.3"

  required_providers {
    google = ">= 3.3"
  }
}

data "google_project" "project" {}

data "google_service_account" "bleeper_run" {
  account_id = "bleeper-run"
}

data "google_sql_database_instance" "bleeper_db" {
  name = "${var.env}-bleeper"
}

data "google_compute_network" "vpc" {
  name = "${var.env}-vpc"
}

data "google_compute_subnetwork" "subnet" {
  name = "${var.env}-vpc-subnet"
}

resource "google_storage_bucket" "bleeper_storage" {
  name     = "${var.env}-bleeper-storage"
  location = "US"

  cors {
    origin          = ["*"]
    method          = ["GET", "HEAD", "PUT", "POST", "DELETE"]
    response_header = ["*"]
    max_age_seconds = 3600
  }
}

resource "google_cloud_run_v2_service" "run_service" {
  name         = "${var.env}-${var.service}"
  location     = var.location
  launch_stage = "BETA"

  template {
    service_account = data.google_service_account.bleeper_run.email

    scaling {
      min_instance_count = local.min_count[var.env]
      max_instance_count = local.max_count[var.env]
    }

    containers {
      image = "gcr.io/${var.env}-selki/${var.service}:${var.revision}"

      env {
        name  = "GCP_LOCATION"
        value = var.location
      }

      env {
        name  = "NODE_ENV"
        value = var.env == "prod" ? "production" : "development"
      }

      env {
        name  = "SKIP_2FA"
        value = var.env == "prod" ? 0 : 1
      }

      env {
        name  = "DATABASE_HOST"
        value = data.google_sql_database_instance.bleeper_db.private_ip_address
      }

      env {
        name  = "INSTANCE_NAME"
        value = "${var.env}-bleeper"
      }

      env {
        name  = "DATABASE_CLIENT"
        value = "postgres"
      }

      env {
        name  = "DATABASE_PORT"
        value = 5432
      }

      env {
        name  = "DATABASE_NAME"
        value = "bleeper"
      }

      env {
        name  = "DATABASE_USERNAME"
        value = "bleeper"
      }

      env {
        name  = "ENABLE_POSTMARK"
        value = 1
      }

      env {
        name  = "GCS_MEDIA_BUCKET_NAME"
        value = "${var.env}-bleeper-storage"
      }

      env {
        name  = "ELASTICSEARCH_URL"
        value = local.es_host[var.env]
      }

      env {
        name  = "ELASTICSEARCH_INDEX"
        value = local.es_index[var.env]
      }

      env {
        name  = "CLICKUP_NOTIFICATIONS_VIEW"
        value = local.clickup_notification_views[var.env]
      }

      env {
        name = "DATABASE_PASSWORD"
        value_source {
          secret_key_ref {
            secret  = "bleeper-DATABASE_PASSWORD"
            version = "latest"
          }
        }
      }

      env {
        name = "SESSION_SECRET"
        value_source {
          secret_key_ref {
            secret  = "bleeper-SESSION_SECRET"
            version = "latest"
          }
        }
      }

      env {
        name = "ELASTICSEARCH_KEY"
        value_source {
          secret_key_ref {
            secret  = "bleeper-ELASTICSEARCH_KEY"
            version = "latest"
          }
        }
      }

      env {
        name = "BURLAK_API_TOKEN"
        value_source {
          secret_key_ref {
            secret  = "bleeper-BURLAK_API_TOKEN"
            version = "latest"
          }
        }
      }

      env {
        name = "POSTMARK_API_TOKEN"
        value_source {
          secret_key_ref {
            secret  = "bleeper-POSTMARK_API_TOKEN"
            version = "latest"
          }
        }
      }

      env {
        name = "CLICKUP_API_TOKEN"
        value_source {
          secret_key_ref {
            secret  = "bleeper-CLICKUP_API_TOKEN"
            version = "latest"
          }
        }
      }

      env {
        name = "GCS_ACCESS_KEY_ID"
        value_source {
          secret_key_ref {
            secret  = "bleeper-GCS_ACCESS_KEY_ID"
            version = "latest"
          }
        }
      }

      env {
        name = "GCS_SECRET_ACCESS_KEY"
        value_source {
          secret_key_ref {
            secret  = "bleeper-GCS_SECRET_ACCESS_KEY"
            version = "latest"
          }
        }
      }

      volume_mounts {
        mount_path = "/cloudsql"
        name       = "cloudsql"
      }
    }

    volumes {
      name = "cloudsql"

      cloud_sql_instance {
        instances = ["${local.project}:${var.location}:${local.hostname[var.env]}"]
      }
    }

    vpc_access {
      network_interfaces {
        network    = "projects/${var.env}-selki/global/networks/${data.google_compute_network.vpc.name}"
        subnetwork = "projects/${var.env}-selki/regions/${var.location}/subnetworks/${data.google_compute_subnetwork.subnet.name}"
      }
      egress = "PRIVATE_RANGES_ONLY"
    }
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }
}

resource "google_cloud_run_v2_service_iam_member" "run_all_users" {
  name     = google_cloud_run_v2_service.run_service.name
  location = google_cloud_run_v2_service.run_service.location
  role     = "roles/run.invoker"
  member   = "allUsers"
}