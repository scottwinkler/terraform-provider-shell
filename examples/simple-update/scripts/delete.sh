#!/bin/bash
IN=$(cat)
id=$(echo $IN | jq -r .id)
rm ${id}.json