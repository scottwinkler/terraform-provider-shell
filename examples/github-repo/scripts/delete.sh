#!/bin/bash

# read from STDIN to get the repo identifier
IN=$(cat)
full_name=$(echo $IN | jq -r .full_name)

# do DELETE
curl \
  --header "Authorization: token $OAUTH_TOKEN" \
  --header "Accept: application/vnd.github.v3+json" \
  --request DELETE \
  https://api.github.com/repos/$full_name