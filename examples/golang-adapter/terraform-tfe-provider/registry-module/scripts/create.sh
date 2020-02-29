#!/bin/bash
../../../modules/golang/linux -name=registry-module -command=create
cat state.json >&3
rm state.json