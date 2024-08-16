packer {
  required_plugins {
    hostmgr = {
     version = ">= 0.25.1"
     source  = "github.com/Automattic/hostmgr"
    }
  }
}

variable vm_name {
  type    = string
}

variable vm_username {
  type    = string
}

variable vm_password {
  type    = string
}

source "hostmgr-builder" "macos-image" {
  source_image = "macos-13.5.2"
  destination_image = "${var.vm_name}"

  ssh_username = "${var.vm_username}"
  ssh_password = "${var.vm_password}"
  ssh_port     = 22
}

build {
  sources = [
    "source.hostmgr-builder.macos-image"
  ]

  provisioner "hostmgr-provisioner" {
#    enable_passwordless_sudo = true

#    homebrew_dependencies = [
#      "python"
#    ]
  }

#  provisioner "ansible" {
#    playbook_file = "example/macos-playbook.yml"
#    use_proxy = false
#  }

  post-processor "hostmgr-uploader" {}

}
