

# runs the "whoami" command and returns user
data "shell_script" "user" {
    lifecycle_commands {
        read = <<-EOF
            echo "{\"user\": \"$(whoami)\"}"
        EOF
    }
}

output "user" {
    value = data.shell_script.user.output["user"]
}

# gets the weather as a data source
data "shell_script" "weather" {
  lifecycle_commands {
    read = <<-EOF
        echo "{\"SanFrancisco\": \"$(curl wttr.in/SanFrancisco?format="%l:+%c+%t")\"}"
    EOF
  }
}

output "data_weather" {
  value = data.shell_script.weather.output["SanFrancisco"]
}

# gets the weather as a resource
resource "shell_script" "weather" {
  lifecycle_commands {
    create = <<-EOF
            echo "{\"London\": \"$(curl wttr.in/London?format="%l:+%c+%t")\"}"  > state.json
            cat state.json
        EOF
    delete = "rm state.json"
  }
}

output "weather" {
  value = shell_script.weather.output["London"]
}
