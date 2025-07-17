variable "env" {
  default = "dev"
}

variable "service" {
  default = "gophish"
}

variable "location" {
  default = "us-central1"
}

variable "zone" {
  default = "us-central1-a"
}

variable "FORCE_RESOURCES" {
  description = "Force creation of resources regardless of environment"
  type        = bool
  default     = false
}