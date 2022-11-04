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

usage() {
  echo "Usage: $0: [-p] [remote] [branch]"
  echo " -p  Increment patch number (0.0.x) intead of minor number (0.x.0)"
  echo " [remote] Remote name ('origin' by default)"
  echo " [branch] Target branch ('main' by default)"
}

INCREMENT_INDEX=1

while getopts p opt; do
  case $opt in
    p)
      INCREMENT_INDEX=2
      ;;
    h|?)
      usage
      exit 2
      ;;
  esac
done

shift $((OPTIND-1))
REMOTE="${1:-origin}"
TARGET_BRANCH="${2:-main}"

assert_clean_checkout() {
  if [ "$(git status --porcelain)" ]; then
    echo "Checkout is not clean"
    git status
    exit 1
  fi
}

join_version() {
  local IFS="."
  echo "$*"
}

increment_version() {
  local parts

  IFS="." read -r -a parts < VERSION

  for i in "${!parts[@]}"; do
    if [ $i -eq $INCREMENT_INDEX ]; then
      ((parts[$i]++))
    elif [ $i -gt $INCREMENT_INDEX ]; then
      parts[$i]=0
    fi
  done

  join_version "${parts[@]}" | tee VERSION
}

commit_has_merged() {
  local commit="$1"

  git fetch $REMOTE

  if ! git merge-base --is-ancestor "$commit" $REMOTE/$TARGET_BRANCH; then
    return 1
  fi

  return 0
}

wait_for_commit() {
  local commit="$1"

  echo "Waiting for commit $commit to merge..."

  until commit_has_merged "$commit"; do
    sleep 5
    echo "..."
  done

  echo "Commit $commit has merged"
}

assert_clean_checkout
git pull --rebase
assert_clean_checkout

if [ "$(git rev-list $REMOTE/$TARGET_BRANCH..)" ]; then
    echo "Local commits detected:"
    git log --oneline $REMOTE/$TARGET_BRANCH..
    echo "Aborting"
    exit 1
fi

make install-tools

version="$(increment_version)"
branch="version/$version"
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

echo "Creating temporary local branch $branch"
git checkout -b "$branch"

echo "Committing VERSION and CHANGELOG.md"
git add VERSION CHANGELOG.md
git commit --message="version: $version"

echo "Creating signed version tag $tag"
git tag --sign --message="version: $version" "$tag" HEAD

echo "Pushing merge request (will merge when pipeline succeeds)"
git push \
  -o merge_request.create \
  -o merge_request.target=$TARGET_BRANCH \
  -o merge_request.merge_when_pipeline_succeeds \
  --set-upstream $REMOTE "$branch"

wait_for_commit "$(git rev-parse HEAD)"

echo "Pushing tag"
git push --tags origin "$tag"

echo "Checking out main and removing temporary branch"
git checkout main
git pull --rebase
git branch -D "$branch"
