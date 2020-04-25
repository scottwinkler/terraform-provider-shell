#!/bin/bash
id=$RANDOM
/bin/cat <<END >$id.json
  {"id": "$id", "description": "$DESCRIPTION"}
END
cat $id.json