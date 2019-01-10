import argparse, sys, json, os, importlib, inspect
from abc import ABCMeta, abstractmethod
from resources import *

# example usage: python3 main.py --name=TestResource --module=test_resource  --command=read --state='{"hello":"world"} && cat state.json >&3'
def main():
    args = sys.argv
    parser = argparse.ArgumentParser()
    parser.add_argument('--module',default='resources',help='The name of the python module to import from')
    parser.add_argument('--name',help='The name of the resource class')
    parser.add_argument('--command',help='The command to execute. One of: create|read|update|delete')
    parser.add_argument('--state',default='{}',help='The previous state of the resource')
    args = parser.parse_args()
    module = args.module
    resource_name = args.name
    command = args.command
    validate_command(command)
    raw_state = args.state
    state = parse_state(raw_state)
    environment = os.environ
    m = globals()[module]
    validate_not_none(m)
    c = getattr(m,resource_name)
    validate_not_none(c)
    resource = initialize_resource(c,environment,state)
    lifecycle_method = getattr(resource,command)
    lifecycle_method()
    new_state = getattr(resource,'state')
    sys_print(new_state)

def sys_print(data):
    file = open('state.json','w')
    file.write(json.dumps(data))
    file.close()

def initialize_resource(c,environment,state):
    param_count = len(inspect.signature(c.__init__).parameters)
    resource = None
    if param_count == 2:
        resource = c(environment)
    elif param_count == 3:
        resource = c(environment,state)
    return resource

def validate_command(command):
    if not command in ['create','read','update','delete']:
        raise ValueError

def validate_not_none(m):
    if m is None:
        raise ValueError

def parse_state(raw):
    data = {}
    if raw is not None:
        try:
            data = json.loads(raw)
        except json.JSONDecodeError:
            print("Error reading json input")
    return data

# Entry point
main()
