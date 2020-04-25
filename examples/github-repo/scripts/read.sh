#!/bin/bash

# read from STDIN to get the repo identifier
IN=$(cat)
full_name=$(echo $IN | jq -r .full_name)

# do GET
curl \
  --header "Authorization: token $OAUTH_TOKEN" \
  --header "Accept: application/vnd.github.v3+json" \
  --request GET \
  https://api.github.com/repos/$full_name
