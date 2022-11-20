#!/bin/bash

new="$1"

current=$(git describe --tags --abbrev=0)
releaseNote=$(git log --pretty="%s (%an) - %h" "${current}"..HEAD)

commitMsg="New tag: ${new}

${releaseNote}"

echo "${commitMsg}"

git tag -a "${new}" -m "${commitMsg}"
