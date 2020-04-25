#!/bin/bash
IN=$(cat)
id=$(echo $IN | jq -r .id)
/bin/cat <<END >$id.json
  {"id": "$id", "description": "$DESCRIPTION"}
END