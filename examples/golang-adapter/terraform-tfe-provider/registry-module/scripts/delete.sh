#!/bin/bash
state=$(cat)
echo $state > state.json
../../../modules/golang/linux -name=registry-module -command=delete -state=state.json
cat state.json >&3
rm state.json