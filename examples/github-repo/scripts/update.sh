#!/bin/bash

# read from STDIN to get the repo identifier
IN=$(cat)
full_name=$(echo $IN | jq -r .full_name)

# make payload for the PATCH request
/bin/cat <<END >payload.json
{
  "name": "$NAME",
  "description": "$DESCRIPTION",
  "homepage": "https://github.com",
  "private": false,
  "has_issues": true,
  "has_projects": true,
  "has_wiki": true
}
END

# do PATCH
curl \
  --header "Authorization: token $OAUTH_TOKEN" \
  --header "Accept: application/vnd.github.v3+json" \
  --data @payload.json \
  --request PATCH \
  https://api.github.com/repos/$full_name

# cleanup
rm payload.json

# do GET - this will be what gets saved to state
curl \
  --header "Authorization: token $OAUTH_TOKEN" \
  --header "Accept: application/vnd.github.v3+json" \
  --request GET \
  https://api.github.com/repos/$full_name