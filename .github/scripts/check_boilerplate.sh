#!/usr/bin/env bash

set -euo pipefail

# operating modes
fix=0
all=0

# counters
errors=0
fixed=0

# ignore these files
ignore_patterns=("**/.gitignore" "go.mod" "go.sum" ".github/*" "LICENSE" "NOTICE" ".*" "*.output" "*.static" "*.long")

# current year, used for replacing $year in the boilerplate
current_year=$(date +"%Y")

# pre-existing modified files should have both the old and new pattern
# new files should have the new pattern
old_pattern="Copyright [0-9,-]+ Weaveworks Ltd."

if [ -n "$GITHUB_PR_NUMBER" ]; then
  # if this is a PR, check for correct copyright year
  new_pattern="Copyright ${current_year} bootloose authors" 
else
  # otherwise accept any year
  new_pattern="Copyright [0-9,-]+ bootloose authors" 
fi

# unix timestamp for 2023-08-30 - original repo was changed to readonly on 2023-08-29
# and the new repo was created on 2023-08-31 so anything created after 2023-08-30
# is a new file 
switch_timestamp=1693395563

if command -v ggrep >/dev/null 2>&1; then
  EGREP="ggrep"
elif command -v egrep >/dev/null 2>&1; then
  EGREP="egrep"
else
  EGREP="grep -E"
fi

# is_pre_existing returns true if the file was created before the switch
is_pre_existing() {
  local filename=$1

  file_timestamp=$(git log --follow --pretty=format:%at -- "$filename" | tail -n 1)
  if [ "$file_timestamp" -lt "$switch_timestamp" ]; then
    return 0  # true
  fi
  return 1  # false
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
  # Check that both the old pattern and new pattern are present
  ${EGREP} -q "$old_pattern" "$filename" && ${EGREP} -q "$new_pattern" "$filename" && return 0
  return 1
}

check_new_file() {
  local filename=$1
  local ret=1

  ${EGREP} -q "$new_pattern" "$filename" && {
    ${EGREP} -Lq "$old_pattern" "$filename" && echo "Warning: a new file contains the old copyright: $filename"
    ret=0
  }

  return $ret
}

inject_boilerplate() {
  local target_file=$1
  local boilerplate_file=$2
  local line_number=$3
  local comment_char=$4
  local begin_comment=$5
  local end_comment=$6

  if [ -n "$comment_char" ]; then
    boilerplate=$(sed "s|^|$comment_char|" "$boilerplate_file")
  else
    boilerplate=$(cat "$boilerplate_file")
  fi

  boilerplate=${boilerplate//\$year/$current_year}

  [ -n "$begin_comment" ] && boilerplate="${begin_comment}\n${boilerplate}"
  [ -n "$end_comment" ] && boilerplate="${boilerplate}\n${end_comment}"
  
  ed_script=$(printf "%si\n%s\n.\nw\nq\n" "$line_number" "$boilerplate")

  echo "$ed_script" | ed -s "$target_file" > /dev/null 2>&1 || {
    echo "Running ed failed for file: $target_file"
    echo "Script content:"
    echo "$ed_script"
  }
}

add_boilerplate() {
  local filename=$1
  local boilerplate=$2

  grep -q "Copyright" "$filename" && {
    echo "File already has a copyright notice: $filename" >&2
    return 1
  }

  case "${filename##*/}" in
    *.go)           inject_boilerplate "$filename" "$boilerplate" 1 "" "/*" "*/"$'\n' && return 0 ;;
    *.cmd|*.yml|*.yaml) inject_boilerplate "$filename" "$boilerplate" 1 "# "              && return 0 ;; 
    *.bash|*.sh)      inject_boilerplate "$filename" "$boilerplate" 2 "# " $'\n'        && return 0 ;;
    *.md|*.html)      inject_boilerplate "$filename" "$boilerplate" 1 "" "<!--" "-->"   && return 0 ;;
  esac

  case "$(basename "$filename")" in
    Dockerfile|Makefile) inject_boilerplate "$filename" "$boilerplate" 1 "# " && return 0 ;;
  esac

  echo "Unsupported filetype for fixing in file: $filename" >&2
  return 1
}

while getopts ":afh" opt; do
  case $opt in
    f) fix=1 ;;
    a) all=1 ;;
    h)  echo "Usage: $0 [options]"
        echo "Options:"
        echo "  -a    Check all files, not just the ones changed in this commit/PR."
        echo "  -f    Fix mode. Adds missing copyright boilerplates."
        echo "  -h    Display this help message."
        exit 0
        ;;
    \?) echo "Invalid option: -$OPTARG" >&2; exit 1 ;;
  esac
done

if [[ $all == 1 ]]; then
  # Get the list of all files in the repo
  files=$(git ls-files)
elif [ -n "$GITHUB_PR_NUMBER" ]; then
  # Get the list of files changed in this PR
  files=$(gh pr view "$GITHUB_PR_NUMBER" --json files --jq '.files.[].path')
else
  # Get the list of files changed in this branch
  local mergeBase="$(git merge-base HEAD main)"
  files=$(git diff --name-only HEAD "$mergeBase")
fi

for filename in $files; do
  if [ ! -f "$filename" ]; then
    # The file was removed, no need to check it.
    continue
  fi

  # reset the mode flag, used to determine which boilerplate to use
  mode=""

  # check if the file should be ignored
  skip=0
  # need to disable SC2053 because need to use the globbing
  # shellcheck disable=SC2053
  for ignore in "${ignore_patterns[@]}"; do
    [[ $filename != $ignore ]]  || { skip=1; break; }
  done
  [ $skip -eq 0 ] || continue

  # check if the file was created before the switch
  if is_pre_existing "$filename"; then
    # if file isn't modified after the switch, skip it
    is_modified_after_switch "$filename" || continue
    mode="old"
  else
    mode="new"
  fi

  check_func="check_${mode}_file"
  ${check_func} "$filename" && continue

  ((errors++))
  echo "Missing .copyright-boilerplate.${mode} from file: $filename" >&2

  if [ -n "$GITHUB_RUN_ID" ]; then
    github_annotate "$filename"
  fi

  if [[ $fix == 1 ]]; then
    echo "Fixing $filename"
    add_boilerplate "$filename" ".copyright-boilerplate.${mode}" && {
      ((fixed++))
      git add "$filename"
    }
  fi
done

if [[ $errors -gt 0 ]]; then
  echo "$errors files are missing or have incorrect copyright boilerplate" >&2
  if [[ $fix == 1 ]]; then
    echo "Fixed $fixed files"
  fi
  exit 1
fi
