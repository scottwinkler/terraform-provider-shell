#!/bin/bash
IN=$(cat)
id=$(echo $IN | jq -r .id)
cat ${id}.json