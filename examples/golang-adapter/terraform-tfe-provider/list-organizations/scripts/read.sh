#!/bin/bash
../../../modules/golang/linux -name=organizations -command=read
cat state.json >&3
rm state.json