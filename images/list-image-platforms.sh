#!/usr/bin/env sh
# SPDX-FileCopyrightText: 2023 bootloose authors
# SPDX-License-Identifier: Apache-2.0

set -eu

image_name="$1"

if [ -z "$image_name" ]; then
  echo "Usage: $0 <image_name>"
  exit 1
fi

# Builds a CSV list of machine image's upstream container's available platforms.
# To be used for GitHub release workflow matrix.

fetch_platforms() {
  image=$1
  repo="${image%:*}"
  tag="${image#*:}"

  case "$repo" in
  */*) ;; # repo name has a prefix already
  *) repo="library/$repo" ;;
  esac

  url="https://hub.docker.com/v2/repositories/$repo/tags/$tag/?page_size=20"

  # outputs linux/arm64,linux/amd64,linux/arm/v7
  curl -s "$url" \
    | jq -r '[.images[] | select(.os=="linux" and (.architecture | test("amd64|arm64|arm"))) | "\(.os)/\(.architecture)\(if .variant then "/" + .variant else "" end)"] | join(",")'
}

upstream_image=$(grep -m 1 'FROM' "${image_name}/Dockerfile" | cut -d ' ' -f2)
fetch_platforms "$upstream_image"
