resource "shell_script" "state_emitted_in_create" {
  lifecycle_commands {
    create = "echo \"I'm doing state the right way\"; echo '{\"size\": \"2\"}'"
    delete = "echo Doing nothing lol"
  }

  environment = {
    foo = "bar"
  }
}

resource "shell_script" "state_not_emitted_in_create" {
  lifecycle_commands {
    create = "echo \"I do something but I don't say what\""
    read   = "echo \"I emit state in the read handler\"; echo '{\"size\": \"2\"}'"
    delete = "echo Doing nothing lol"
  }

  environment = {
    foo = "bar"
  }
}
