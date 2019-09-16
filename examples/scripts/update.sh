#!/bin/bash
echo "updating..."
echo "writing some error" >&2

IN=$(cat)
echo "stdin: ${IN}" #the old state

#business logic
/bin/cat <<END >ex.json
  {"commit_id": "b8f2b8b", "environment": "$yolo", "tags_at_commit": "sometags", "project": "someproject", "current_date": "09/10/2014", "version": "someversion"}
END