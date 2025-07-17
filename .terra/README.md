# Gophish Infrastructure

This directory contains Terraform configuration for managing the Gophish infrastructure in the prod-selki GCP project.

## Resources

- **Compute Disk**: `prod-gophish-host` - A 50GB standard persistent disk in us-central1-a

## Prerequisites

1. Ensure you're authenticated with GCP:
   ```bash
   gcloud auth login
   gcloud config set project prod-selki
   ```

2. Initialize Terraform:
   ```bash
   cd infra
   terraform init
   ```

## Usage

1. Review the planned changes:
   ```bash
   terraform plan
   ```

2. Apply the configuration (requires approval):
   ```bash
   terraform apply
   ```

## Import Existing Resources

To import the existing disk:
```bash
terraform import google_compute_disk.gophish_host projects/prod-selki/zones/us-central1-a/disks/prod-gophish-host
```

## State Storage

Terraform state is stored in GCS bucket `prod-selki-terraform-state` with prefix `gophish`.