#!/bin/bash

# Builds a JSON matrix of all Docker images and their supported platforms
# for to be used in the GitHub Actions workflow matrix

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
first_image=true
for dockerfile in images/**/Dockerfile; do
  image_name=$(basename $(dirname $dockerfile))
  upstream_image=$(grep -m 1 'FROM' $dockerfile | cut -d ' ' -f2)
  platforms=$(fetch_platforms $upstream_image)

  if [ "$first_image" = true ]; then
    first_image=false
  else
    items+=","
  fi  
  items+="{\"image\": \"$image_name\",\"platforms\":["
  first_platform=true
  for platform in $platforms; do
    if [ "$first_platform" = true ]; then
      first_platform=false
    else
      items+=","
    fi
    items+="\"$platform\""
  done
  items+="]}"
done

echo "[$items]"