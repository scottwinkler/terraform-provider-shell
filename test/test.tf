provider "shell" {}

data "shell_script" "test" {
  read = <<EOF
  echo '{"commit_id": "b8f2b8b", "environment": "test", "tags_at_commit": "sometags", "project": "someproject", "current_date": "09/10/2014", "version": "someversion"}';
 
EOF

  working_directory = "./tmp"

  environment = {
    yolo = "yolo"
  }
}

output "commit_id" {
  value = "${data.shell_script.test.output["commit_id"]}"
}

resource "shell_script" "test" {
  create = <<EOF
  /bin/cat <<END >ex.json
  {"commit_id": "b8f2b8b", "environment": "$yolo", "tags_at_commit": "sometags", "project": "someproject", "current_date": "09/10/2014", "version": "someversion"}
END
EOF

  read = <<EOF
    cat ex.json
EOF

  delete = <<EOF
  rm -rf ex.json
EOF

  working_directory = "./tmp"

  environment = {
    yolo = "yolo"
  }
}

output "environment" {
  value = "${shell_script.test.output["environment"]}"
}
