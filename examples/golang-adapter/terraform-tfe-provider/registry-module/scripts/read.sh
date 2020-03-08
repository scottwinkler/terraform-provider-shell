#!/bin/bash
state=$(cat)
echo $state > state.json
../../../modules/golang/linux -name=registry-module -command=read -state=state.json
cat state.json
rm state.json
