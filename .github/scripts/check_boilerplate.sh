#!/usr/bin/env bash
# SPDX-FileCopyrightText: 2023 bootloose authors
# SPDX-License-Identifier: Apache-2.0

set -xeuo pipefail

# setup variable defaults s to avoid unbound variable errors
GITHUB_REPOSITORY=${GITHUB_REPOSITORY:-"k0sproject/bootloose"}
GITHUB_PR_NUMBER=${GITHUB_PR_NUMBER:-""}
GITHUB_RUN_ID=${GITHUB_RUN_ID:-""}
DEBUG=${DEBUG:-""}

[ -n "$DEBUG" ] && set -x

# operating modes
fix=0
all=0
dry=0

# counters
errors=0
fixed=0

# ignore these files
ignore_patterns=("**/.gitignore" "go.mod" "go.sum" ".github/*" "LICENSE" "NOTICE" ".*" "*.output" "*.static" "*.long")
for filename in pkg/docker/*; do
  if grep -q "Kubernetes Authors" "$filename"; then
    ignore_patterns+=("$filename")
  fi
done

# all boilerplates should include the SPDX license identifier
spdx_identifier="SPDX-License-Identifier: Apache-2.0"

# unix timestamp for 2023-08-30 - original repo was changed to readonly on 2023-08-29
# and the new repo was created on 2023-08-31 so anything created after 2023-08-30
# is a new file 
switch_timestamp=1693395563

declare -r new_holder='bootloose authors'
declare -ar original_holders=(
  'Weaveworks Ltd.'
  'The Kubernetes Authors.'
)

error_echo() {
  echo "$@" >&2
}

debug_echo() {
  [ -n "$DEBUG" ] && error_echo "$@"
}

panic() {
  error_echo "$@"
  exit 1
}

file_introduction_date() {
  git log --diff-filter=A --author-date-order --format=%ad --date="${2-unix}" -- "$1" | tail -n1
}

file_introduction_year() {
  file_introduction_date "$1" format:%Y
}

file_first_modified_date() {
  git log --author-date-order --format=%ad --date="${2-unix}" --after=$switch_timestamp -- "$1" | tail -n1
}

file_first_modified_year() {
  file_first_modified_date "$1" format:%Y
}

spdx_copyright_pattern() {
  local filename=$1
  local mode=$2
  local holder

  case "$mode" in
    old)
      if [ $# -gt 2 ] && [ "$3" != "${original_holders[0]}" ] ; then
        year="[1-9][0-9][0-9][0-9]"
        holder="$3"
      else
        year=$(file_introduction_year "$filename")
        holder="${original_holders[0]}"
      fi
      ;;
    new)
      year=$(file_first_modified_year "$filename")
      holder="$new_holder"
      ;;
    ?) panic "Invalid mode: $mode" ;;
  esac

  echo -n "SPDX-FileCopyrightText: ${year} ${holder}"
}

require_spdx_identifier() {
  local filename=$1

  grep -q "$spdx_identifier" "$filename" || {
    error_echo "${filename}: Missing SPDX license identifier"
    return 1
  }
  return 0
}

require_spdx_pattern_old() {
  local filename=$1
  local holder found=0

  for holder in "${original_holders[@]}"; do
    if grep -q "$(spdx_copyright_pattern "$filename" old "$holder")" "$filename"; then
      found=1
      break
    fi
  done

  [ $found -eq 1 ] || {
    error_echo "${filename}: Missing old SPDX license pattern"
    return 1
  }
}

require_spdx_pattern_new() {
  local filename=$1

  grep -q "$(spdx_copyright_pattern "$filename" new)" "$filename" || {
    error_echo "${filename}: Missing new SPDX license pattern"
    return 1
  }
  return 0
}

# is_pre_existing returns true if the file was created before the switch
is_pre_existing() {
  local filename=$1
  local since

  since=$(file_introduction_date "$filename") || return 2
  if [ "$since" -lt "$switch_timestamp" ]; then
    return 0
  else
    return 1
  fi
}

# is_modified_after_switch returns true if the file was modified after the switch
is_modified_after_switch() {
  local filename=$1

  # Get the timestamp of the most recent commit for this file
  latest_timestamp=$(git log -n 1 --pretty=format:%at -- "$filename")

  if [ "$latest_timestamp" -gt "$switch_timestamp" ]; then
    return 0  # true
  fi
  return 1  # false
}

github_annotate() {
  local filename=$1

  gh api "repos/${GITHUB_REPOSITORY}/check-runs/${GITHUB_RUN_ID}/annotations" \
    -X POST \
    -F path="$filename" \
    -F start_line=1 \
    -F end_line=1 \
    -F annotation_level="failure" \
    -F message="Missing or incorrect copyright boilerplate" \
    || echo "Failed to add a github annotation to file: $filename"
}

check_old_file() {
  local filename=$1
  # Check that identifier and both the old pattern and the new pattern are present
  if require_spdx_identifier "$filename" && \
    require_spdx_pattern_old "$filename" && \
    require_spdx_pattern_new "$filename"; then
    return 0
  fi
  return 1
}

check_new_file() {
  local filename=$1
  local holder

  # Check that identifier and the new pattern are present
  if require_spdx_identifier "$filename" && require_spdx_pattern_new "$filename"; then
    # warn if an old pattern is present
    for holder in "${original_holders[@]}"; do
      if grep -i "spdx" "$filename" | grep -iq "$holder"; then
        error_echo "${filename}: Warning: a newly introduced file contains an old copyright"
        break
      fi
    done

    return 0
  fi

  return 1
}

inject_boilerplate() {
  local filename=$1
  local mode=$2
  local line_number=$3
  local comment_char=$4
  # optional arguments
  local begin_comment=${5:-}
  local end_comment=${6:-}

  case "$mode" in
    old) 
      boilerplate="${comment_char}$(spdx_copyright_pattern "$filename" "old")"$'\n'
      boilerplate+="${comment_char}$(spdx_copyright_pattern "$filename" "new")"
      ;;
    new) 
      boilerplate="${comment_char}$(spdx_copyright_pattern "$filename" "new")"
      ;;
    *) panic "Invalid mode: $mode" ;;
  esac

  boilerplate+=$'\n'"${comment_char}${spdx_identifier}"

  if [ -n "$begin_comment" ]; then
    boilerplate="${begin_comment}"$'\n'"${boilerplate}"
  fi

  if [ -n "$end_comment" ]; then
    if [ "$end_comment" == $'\n' ]; then
      boilerplate+=$'\n'
    else
      boilerplate+=$'\n'"${end_comment}"
    fi
  fi

  ed_script=$(printf "%si\n%b\n.\nw\nq\n" "$line_number" "$boilerplate")

  if [[ $dry == 1 ]]; then
    echo "${filename} : DRY: ed_script='$ed_script'"
    echo "${filename} : DRY: echo $ed_script | ed -s \"$filename\""
    return 0
  fi

  echo "$ed_script" | ed -s "$filename" > /dev/null 2>&1 || {
    error_echo "${filename}: Running ed failed"
    debug_echo "Script content:"
    debug_echo "$ed_script"
    return 1
  }
}

add_boilerplate() {
  local filename=$1
  local mode=$2

  if grep -q "Copyright" "$filename"; then
    error_echo "${filename}: File already has a copyright notice"
    return 1
  fi

  case "${filename##*/}" in
    *.go)
      inject_boilerplate "$filename" "$mode" 1 "// " "" $'\n' ;;
    *.cmd|*.yml|*.yaml)
      inject_boilerplate "$filename" "$mode" 1 "# " ;; 
    *.bash|*.sh)
      inject_boilerplate "$filename" "$mode" 2 "# " ;;
    *.md|*.html)
      inject_boilerplate "$filename" "$mode" 1 "" "<!--" "-->" ;;
    Dockerfile*|Makefile*)
      inject_boilerplate "$filename" "$mode" 1 "# " ;;
    *)
      error_echo "${filename}: Unsupported filetype for adding boilerplate"
      return 1
      ;;
  esac
}

list_files() {
  if [[ $all == 1 ]]; then
    # Get the list of all files in the repo
    git ls-files
  elif [ -n "$GITHUB_PR_NUMBER" ]; then
    # Get the list of files changed in this PR
    gh pr view "$GITHUB_PR_NUMBER" --json files --jq '.files.[].path'
  else
    # Get the list of files changed in this branch
    local mergeBase
    mergeBase="$(git merge-base HEAD main)"
    git diff --name-only HEAD "$mergeBase"
  fi
}

while getopts ":afhd" opt; do
  case $opt in
    f) fix=1 ;;
    a) all=1 ;;
    d) dry=1 ;;
    h)  echo "Usage: $0 [options]"
        echo "Options:"
        echo "  -a    Check all files, not just the ones changed in this commit/PR."
        echo "  -f    Fix mode. Adds missing copyright boilerplates."
        echo "  -d    Dry run. Don't actually fix anything."
        echo "  -h    Display this help message."
        exit 0
        ;;
    \?) echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
  esac
done

declare -a files

# slightly weird syntax to use while read ... that avoids a subshell
# using a file descriptor, because <<< or < <() do not work on the
# antique version of bash that is on macs.
exec 3< <(list_files)
while read -r filename <&3; do
  if [ ! -f "$filename" ]; then
    # The file was removed, no need to check it.
    continue
  fi

  # check if the file should be ignored
  skip=0
  # need to disable SC2053 because need to use the globbing
  # shellcheck disable=SC2053
  for ignore in "${ignore_patterns[@]}"; do
    [[ $filename != $ignore ]]  || { skip=1; break; }
  done
  [ $skip -eq 0 ] || continue

  files+=("$filename")
done
# close the file descriptor 3
exec 3<&-

for filename in "${files[@]}"; do
  # reset the mode flag, used to determine which boilerplate to use
  mode=""

  # check if the file was created before the switch
  if is_pre_existing "$filename"; then
    # if file isn't modified after the switch, skip it
    is_modified_after_switch "$filename" || continue
    mode="old"
  else
    [ $? -eq 1 ]
    mode="new"
  fi

  check_func="check_${mode}_file"
  if ! ${check_func} "$filename"; then
    ((errors++)) || true # true needed for -e
    error_echo "${filename}: Missing or incorrect copyright boilerplate"

    if [ -n "$GITHUB_RUN_ID" ]; then
      github_annotate "$filename"
    fi

    if [[ $fix == 1 ]]; then
      if add_boilerplate "$filename" "$mode"; then
        ((fixed++)) || true # true needed for -e
        git add "$filename"
        echo "${filename}: Fixed"
      fi
    fi
  fi
done

if [[ $errors -gt 0 ]]; then
  error_echo "* $errors files are missing or have incorrect copyright boilerplate" >&2
  if [[ $fix == 1 ]]; then
    echo "* Fixed $fixed files"
    if [[ $fixed -eq "$errors"  ]]; then
      exit 0
    fi
  fi
  exit 1
fi
