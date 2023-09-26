#!/bin/bash

image_name="$1"

if [ -z "$image_name" ]; then
  echo "Usage: $0 <image_name>"
  exit 1
fi

# Builds a CSV list of machine image's upstream container's available platforms.
# To be used for GitHub release workflow matrix.

fetch_platforms() {
  image=$1
  repo=$(echo $image | cut -d ':' -f 1)
  tag=$(echo $image | cut -d ':' -f 2)

  if [[ ! $repo =~ / ]]; then
    repo="library/$repo"
  fi

  url="https://hub.docker.com/v2/repositories/$repo/tags/$tag/?page_size=20"

  curl -s "$url" \
    | jq -r '.images[] | select(.os=="linux" and (.architecture | test("amd64|arm64|arm"))) | .architecture' \
    | sort | uniq
  }


items=""
upstream_image=$(grep -m 1 'FROM' "${image_name}/Dockerfile" | cut -d ' ' -f2)
platforms=$(fetch_platforms $upstream_image)

first_platform=true

for platform in $platforms; do
  if [ "$first_platform" = true ]; then
    first_platform=false
  else
    list+=","
  fi
  list+="$platform"
done

echo "${list}"
