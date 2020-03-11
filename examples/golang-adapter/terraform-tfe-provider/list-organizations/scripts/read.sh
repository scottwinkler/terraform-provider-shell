#!/bin/bash
../../../modules/golang/linux -name=organizations -command=read
cat state.json
rm state.json
