#!/bin/bash
echo "creating..."
echo "writing some error" >&2

IN=$(cat)
echo "stdin: ${IN}" #the old state, not useful for create step since the old state was empty

#business logic
/bin/cat <<END >ex.json
  {"commit_id": "b8f2b8b", "environment": "$yolo", "tags_at_commit": "sometags", "project": "someproject", "current_date": "09/10/2014", "version": "someversion"}
END
cat ex.json >&3