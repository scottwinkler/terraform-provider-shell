#!/bin/bash
../../../modules/golang/linux -name=registry-module -command=create
cat state.json
rm state.json
