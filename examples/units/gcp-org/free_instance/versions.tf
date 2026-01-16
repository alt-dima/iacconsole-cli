terraform {
  required_version = ">= 1.1"

  backend "gcs" {}

  required_providers {
    google = {
      source = "hashicorp/google"
      version = "5.27.0"
    }
  }
}