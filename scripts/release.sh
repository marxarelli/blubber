#!/bin/bash
#
# Release a new version of Blubber:
#
#  1. Increment value in VERSION (minor by default; pass `-p` to do a
#     patch release).
#  2. Generate CHANGELOG.md using `git chglog`.
#  3. Create a commit for the new version and change log.
#  4. Create a signed version tag.
#  5. Push new commit and version tag.
#
set -o errexit -o nounset -o pipefail

INCREMENT_PLACE=2

while getopts p opt; do
  case $opt in
    p)
      INCREMENT_PLACE=3
      ;;
    h|?)
      echo "Usage: $0: [-p]"
      echo " -p  Increment patch number (0.0.x) intead of minor number (0.x.0)"
      exit 2
      ;;
  esac
done

assert_clean_checkout() {
  if [ "$(git status --porcelain)" ]; then
    echo "Checkout is not clean"
    git status
    exit 1
  fi
}

increment_version() {
  local tmp="$(mktemp)"
  trap 'rm $tmp' RETURN
  awk -F '.' '{ $'"$INCREMENT_PLACE"'++; print $1"."$2"."$3 }' VERSION | tee "$tmp"
  cp "$tmp" VERSION
}

assert_clean_checkout
git pull --rebase
assert_clean_checkout

if [ "$(git rev-list origin/main..)" ]; then
    echo "Local commits detected:"
    git log --oneline origin/main..
    echo "Aborting"
    exit 1
fi

make install-tools

version="$(increment_version)"
tag="v$version"
echo "New version: $version"
echo "New tag: $tag"

git chglog --next-tag "$tag" -o CHANGELOG.md
git diff

read -p "Does everything look ok? [Y/n] " confirmation
confirmation=${confirmation:-Y}

if [ "$confirmation" != "Y" ] && [ "$confirmation" != "y" ]; then
  echo "Aborting"
  exit 1
fi

echo "Creating commit"
git add VERSION CHANGELOG.md
git commit --message="version: $version"

echo "Creating signed version tag $tag"
git tag --sign --message="version: $version" "$tag" HEAD

echo "Pushing version commit"
git push origin HEAD:main

echo "Pushing tag"
git push --tags origin "$tag"
