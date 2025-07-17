// ====================
// Locals
// ====================

locals {
  project = "${var.env}-selki"
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
    prefix = "gophish"
  }

  required_version = ">= 1.3"

  required_providers {
    google = ">= 3.3"
  }
}

data "google_project" "project" {}

data "google_service_account" "gophish_run" {
  account_id = "gophish-run"
}

// ====================
// Resources
// ====================

# Host Disk
resource "google_compute_disk" "gophish_host" {
  count = var.env == "prod" || var.FORCE_RESOURCES ? 1 : 0
  name = "${var.env}-gophish-host"
  type = "pd-ssd"
  zone = var.zone
  size = 50

  image = "https://www.googleapis.com/compute/v1/projects/debian-cloud/global/images/debian-12-bookworm-v20250513"

  physical_block_size_bytes = 4096

  guest_os_features {
    type = "GVNIC"
  }

  guest_os_features {
    type = "SEV_CAPABLE"
  }

  guest_os_features {
    type = "SEV_LIVE_MIGRATABLE_V2"
  }

  guest_os_features {
    type = "UEFI_COMPATIBLE"
  }

  guest_os_features {
    type = "VIRTIO_SCSI_MULTIQUEUE"
  }
}

#
resource "google_compute_address" "gophish_external" {
  count = var.env == "prod" || var.FORCE_RESOURCES ? 1 : 0
  name         = "${var.env}-gophish-external"
  region       = var.location
  address_type = "EXTERNAL"
  network_tier = "PREMIUM"
}

resource "google_compute_address" "gophish_internal" {
  count = var.env == "prod" || var.FORCE_RESOURCES ? 1 : 0
  name         = "${var.env}-gophish-internal"
  region       = var.location
  address_type = "INTERNAL"
  purpose      = "GCE_ENDPOINT"
  network_tier = "PREMIUM"
  subnetwork   = "${var.env}-vpc-subnet"
}

# Host Instance
resource "google_compute_instance" "gophish_host" {
  count = var.env == "prod" || var.FORCE_RESOURCES ? 1 : 0
  name         = "${var.env}-gophish-host"
  machine_type = "e2-medium"
  zone         = var.zone

  tags = ["gophish", "http-server", "https-server"]

  boot_disk {
    auto_delete = false
    device_name = "${var.env}-gophish-host"
    source      = google_compute_disk.gophish_host[0].id
  }

  network_interface {
    network    = "${var.env}-vpc"
    subnetwork = "${var.env}-vpc-subnet"
    network_ip = google_compute_address.gophish_internal[0].address

    access_config {
      nat_ip       = google_compute_address.gophish_external[0].address
      network_tier = "PREMIUM"
    }
  }

  service_account {
    email = data.google_service_account.gophish_run.email
    scopes = [
      "https://www.googleapis.com/auth/devstorage.read_only",
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring.write",
      "https://www.googleapis.com/auth/service.management.readonly",
      "https://www.googleapis.com/auth/servicecontrol",
      "https://www.googleapis.com/auth/trace.append",
    ]
  }

  shielded_instance_config {
    enable_integrity_monitoring = true
    enable_secure_boot          = false
    enable_vtpm                 = true
  }

  metadata = {
    enable-osconfig = "TRUE"
  }
}