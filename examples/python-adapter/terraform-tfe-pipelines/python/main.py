import argparse, sys, json, os, importlib, inspect
from abc import ABCMeta, abstractmethod
from resources import *

# example usage: python3 main.py --name=TestResource --module=test_resource  --command=read --state='{"hello":"world"} && cat state.json'
def main():
    args = sys.argv
    parser = argparse.ArgumentParser()
    parser.add_argument('--module',default='test_resource',help='The name of the python module to import from')
    parser.add_argument('--name',help='The name of the resource class')
    parser.add_argument('--command',help='The command to execute. One of: create|read|update|delete')
    parser.add_argument('--state',default='state.json',help='The path to the state file')
    args = parser.parse_args()
    module = args.module
    resource_name = args.name
    command = args.command
    validate_command(command)
    state_path = args.state
    state = parse_state(state_path)
    environment = os.environ
    m = globals()[module]
    c = getattr(m,resource_name)
    resource = initialize_resource(c,environment,state)
    lifecycle_method = getattr(resource,command)
    lifecycle_method()
    new_state = getattr(resource,'state')
    sys_print(state_path,new_state)

def sys_print(path,data):
    #stringify
    processed_data = {}
    print(data)
    for key, value in data.items():
        processed_data[key]=str(value)
    if os.path.exists(path):
        os.remove(path)
    file = open(path,'w')
    file.write(json.dumps(processed_data))
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

def parse_state(path):
    data = {}
    try:
        f = open(path,'r')
        data = json.load(f)
    except json.JSONDecodeError:
        print("Error reading json input")
    except FileNotFoundError:
        print("File not found, using default")
    return data

# Entry point
main()
