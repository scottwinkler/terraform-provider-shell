resource "shell_script" "no_trailing_spaces" {
  lifecycle_commands {
    create = "echo \"I'm doing state the right way\"; echo '{\"size\": \"2\"}'"
    delete = "echo Doing nothing lol"
  }

  environment = {
    foo = "bar"
  }
}

resource "shell_script" "trailing_spaces" {
  lifecycle_commands {
    create = "echo \"How about some trailing spaces?\"; echo '{\"size\": \"2\"} '"
    delete = "echo Doing nothing lol"
  }

  environment = {
    foo = "bar"
  }
}
