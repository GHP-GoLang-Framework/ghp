#!/usr/bin/env bash
# Enforce the PR issues policy: a PR closes at least one GitHub issue (several
# are allowed), each linked issue carries exactly one version:* label, and the
# release uses the highest bump among them. Dependabot PRs are exempt.
#
# --self-test runs the same validation logic against in-memory fixtures, so
# the rules are tested without gh/network access.
set -euo pipefail

DEPENDABOT=dependabot
CLOSE_KEYWORDS='closes?|closed|fixes?|fixed|resolves?|resolved'

AUTHOR=''
BODY=''
LINKED=()
declare -A LABELS=()
ERRORS=()
BUMP=''

# close_numbers prints the issue numbers closed by explicit keywords in BODY,
# numerically sorted and deduplicated. "Refs #12" is a mention and is not a
# close keyword.
close_numbers() {
  printf '%s\n' "$BODY" \
    | grep -Ei "^[[:space:]>]*(${CLOSE_KEYWORDS})[[:space:]]+#[0-9]+([[:space:]]|$)" \
    | grep -oE '#[0-9]+' \
    | tr -d '#' \
    | sort -nu \
    || true
}

# compute_bump selects the highest version label among all linked issues;
# bumps are ranked, never summed.
compute_bump() {
  BUMP=''
  local best=0 rank label number
  for number in "${!LABELS[@]}"; do
    for label in ${LABELS[$number]}; do
      case "$label" in
        version:major) rank=3 ;;
        version:minor) rank=2 ;;
        version:patch) rank=1 ;;
        *) continue ;;
      esac
      if (( rank > best )); then
        best=$rank
        BUMP=${label#version:}
      fi
    done
  done
}

# validate fills ERRORS (and BUMP); it never fails the script by itself.
validate() {
  ERRORS=()
  if [[ "$AUTHOR" == *"$DEPENDABOT"* ]]; then
    BUMP=patch
    return
  fi

  local linked_list explicit_str linked_str number label
  local -a bumps explicit=()

  linked_list=$(printf '%s\n' "${LINKED[@]}" | sort -nu)
  mapfile -t explicit < <(close_numbers)

  if [[ -z "$linked_list" ]]; then
    ERRORS+=('expected at least one linked issue, GitHub reports none')
  fi
  if (( ${#explicit[@]} == 0 )); then
    ERRORS+=("the PR body must contain at least one 'Closes #N' line")
  else
    explicit_str=$(printf '%s\n' "${explicit[@]}" | paste -sd, -)
    linked_str=$(printf '%s\n' "$linked_list" | paste -sd, -)
    if [[ "$explicit_str" != "$linked_str" ]]; then
      ERRORS+=("the PR body closes [$explicit_str] but GitHub links [$linked_str]")
    fi
  fi

  for number in $linked_list; do
    bumps=()
    for label in ${LABELS[$number]:-}; do
      case "$label" in
        version:major | version:minor | version:patch) bumps+=("${label#version:}") ;;
      esac
    done
    if (( ${#bumps[@]} != 1 )); then
      ERRORS+=("issue #$number must have exactly one version:* label; found ${#bumps[@]}")
    fi
  done

  compute_bump
}

load_from_gh() {
  local pr_json number
  pr_json=$(gh pr view "$1" --json author,body,closingIssuesReferences)
  AUTHOR=$(jq -r '.author.login // ""' <<<"$pr_json")
  BODY=$(jq -r '.body // ""' <<<"$pr_json")
  mapfile -t LINKED < <(jq -r '.closingIssuesReferences[].number' <<<"$pr_json" | sort -nu)

  LABELS=()
  for number in "${LINKED[@]}"; do
    LABELS[$number]=$(
      gh issue view "$number" --json labels \
        --jq '[.labels[].name | select(startswith("version:"))] | join(" ")'
    )
  done
}

self_test() {
  local failed=0
  local -a pairs
  local pair key value

  check_case() {
    local name=$1 author=$2 body=$3 linked=$4 labels_spec=$5
    local want_errors=$6 want_bump=$7
    local got joined

    AUTHOR=$author
    BODY=$body
    LINKED=()
    [[ -n "$linked" ]] && read -ra LINKED <<<"$linked"
    LABELS=()
    if [[ -n "$labels_spec" ]]; then
      IFS=';' read -ra pairs <<<"$labels_spec"
      for pair in "${pairs[@]}"; do
        key=${pair%%=*}
        value=${pair#*=}
        LABELS[$key]=$value
      done
    fi

    validate
    got=${#ERRORS[@]}
    joined=${ERRORS[*]}
    if (( got != want_errors )); then
      printf 'FAIL %s: expected %d errors, got %d (%s)\n' "$name" "$want_errors" "$got" "$joined"
      failed=1
      return
    fi
    if [[ -n "$want_bump" && "$BUMP" != "$want_bump" ]]; then
      printf 'FAIL %s: expected bump %s, got %s\n' "$name" "$want_bump" "$BUMP"
      failed=1
      return
    fi
    printf 'ok - %s\n' "$name"
  }

  check_case 'one linked versioned issue' 'Shimigui' $'Summary\n\nCloses #112\n' \
    '112' '112=version:patch' 0 patch
  check_case 'several linked issues' 'Shimigui' $'Closes #12\nCloses #13\n' \
    '12 13' '12=version:minor;13=version:patch' 0 minor
  check_case 'major wins over minor and patch' 'Shimigui' $'Closes #12\nCloses #13\nCloses #14\n' \
    '12 13 14' '12=version:patch;13=version:major;14=version:minor' 0 major
  check_case 'no linked issue' 'Shimigui' 'Summary with no issue.' \
    '' '' 2 ''
  check_case $'ref-style "Refs: GHP-123"' 'Shimigui' 'Refs: GHP-123' \
    '' '' 2 ''
  check_case 'body closes a different set' 'Shimigui' 'Closes #7' \
    '9' '9=version:patch' 1 ''
  check_case 'linked issue without a version label' 'Shimigui' 'Closes #12' \
    '12' '12=' 1 ''
  check_case 'one of several issues lacks a label' 'Shimigui' $'Closes #12\nCloses #13\n' \
    '12 13' '12=version:minor;13=' 1 ''
  check_case 'linked issue with two version labels' 'Shimigui' 'Closes #12' \
    '12' '12=version:major version:patch' 1 ''
  check_case 'dependabot is exempt' 'app/dependabot' 'Bumps x/term.' \
    '' '' 0 patch
  check_case 'dependabot[bot] spelling is exempt' 'dependabot[bot]' '' \
    '' '' 0 patch
  check_case 'quoted close line' 'Shimigui' '> Closes #21' \
    '21' '21=version:minor' 0 minor

  return "$failed"
}

main() {
  if [[ "${1:-}" == '--self-test' ]]; then
    if self_test; then return 0; else return 1; fi
  fi
  if (( $# != 1 )); then
    echo 'usage: check-pr-issue.sh <pull-request-number> | --self-test' >&2
    return 2
  fi

  load_from_gh "$1"
  validate
  if (( ${#ERRORS[@]} > 0 )); then
    echo '::error::PR issues policy failed:' >&2
    printf -- '- %s\n' "${ERRORS[@]}" >&2
    echo 'Close at least one issue with "Closes #N", give every linked issue exactly one version:* label, or use a Dependabot PR for dependency updates.' >&2
    return 1
  fi

  if [[ "$AUTHOR" == *"$DEPENDABOT"* ]]; then
    echo 'Dependabot PR: exempt from the issue-link rule; bump = patch.'
  else
    echo "PR closes [${LINKED[*]}]; highest version bump = $BUMP."
  fi
}

main "$@"
