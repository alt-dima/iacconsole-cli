resource "google_compute_instance" "vm_instance" {
  name         = "free-instance"
  machine_type = "e2-micro"

  tags = [ "free-instance" ]

  boot_disk {
    initialize_params {
      //size = 15
      image = "ubuntu-os-cloud/ubuntu-minimal-2204-jammy-v20240430"
    }
  }

  metadata = {
    ssh-keys               = "username:ssh-key"
    block-project-ssh-keys = true
  }

  network_interface {
    network = "default"
    # A default network is created for all GCP projects
    #network = google_compute_network.vpc_network.self_link
    access_config {
    }
  }
}
