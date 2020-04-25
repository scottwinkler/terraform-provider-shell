#!/bin/bash

# make payload for the POST request
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

# do POST
resp=$(curl \
  --header "Authorization: token $OAUTH_TOKEN" \
  --header "Accept: application/vnd.github.v3+json" \
  --request POST \
  --data @payload.json \
  https://api.github.com/user/repos)

# cleanup
rm payload.json

# do GET - this will be what gets saved to state
full_name=$(echo $resp | jq -r .full_name)
echo $full_name
curl \
  --header "Authorization: token $OAUTH_TOKEN" \
  --header "Accept: application/vnd.github.v3+json" \
  --request GET \
  https://api.github.com/repos/$full_name