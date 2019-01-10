provider "shell" {}

resource "shell_script" "test" {
  lifecycle_commands = {
    create = "python3 ${path.module}/python/main.py --name=TestResource --module=test_resource --command=create --state='${jsonencode(zipmap(list("hello"),list("world")))}' && cat ${path.module}/state.json >&3 && rm -rf ${path.module}/state.json"
    read = "python3 ${path.module}/python/main.py --name=TestResource --module=test_resource --command=read --state='${jsonencode(zipmap(list("hello"),list("world")))}' && cat ${path.module}/state.json >&3 && rm -rf ${path.module}/state.json"
    update = "python3 ${path.module}/python/main.py --name=TestResource --module=test_resource --command=update --state='${jsonencode(zipmap(list("hello"),list("world")))}' && cat ${path.module}/state.json >&3 && rm -rf ${path.module}/state.json"
    delete = "python3 ${path.module}/python/main.py --name=TestResource --module=test_resource --command=delete --state='${jsonencode(zipmap(list("hello"),list("world")))}' && cat ${path.module}/state.json >&3 && rm -rf ${path.module}/state.json"
  }
}

data "shell_script" "test" {
  lifecycle_commands = {
    read = "python3 ${path.module}/python/main.py --name=TestResource --module=test_resource --command=read --state='${jsonencode(zipmap(list("hello"),list("world")))}' && cat ${path.module}/state.json >&3 && rm -rf ${path.module}/state.json"
  }
}