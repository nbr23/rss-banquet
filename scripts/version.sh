#!/usr/bin/env bash
set -euo pipefail

number='(0|[1-9][0-9]*)'
identifier='(0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)'
pattern="^v${number}\\.${number}\\.${number}(-${identifier}(\\.${identifier})*)?(\\+[0-9A-Za-z-]+(\\.[0-9A-Za-z-]+)*)?$"
matches=()
while IFS= read -r tag; do
    if [[ "$tag" =~ $pattern ]]; then
        matches+=(--match "$tag")
    fi
done < <(git tag --list)

tag=''
if [[ "${#matches[@]}" -gt 0 ]]; then
    tag="$(git describe --tags --abbrev=0 "${matches[@]}" 2>/dev/null || true)"
fi

if [[ -n "$tag" ]]; then
    count="$(git rev-list --count "${tag}..HEAD")"
    if [[ "$count" == 0 ]]; then
        printf '%s\n' "$tag"
        exit 0
    fi
    base="${tag%%+*}"
else
    base='v0.0.0'
    count="$(git rev-list --count HEAD)"
fi

separator='-'
if [[ "$base" == *-* ]]; then
    separator='.'
fi
revision="$(git rev-parse --short=12 HEAD)"
printf '%s%sdev.%s+g%s\n' "$base" "$separator" "$count" "$revision"
