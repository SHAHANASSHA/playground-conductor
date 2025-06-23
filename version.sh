#!/bin/bash

set -e

LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

echo "Last tag: $LAST_TAG"

COMMITS=$(git log ${LAST_TAG}..HEAD --pretty=format:"%s%n%b" 2>/dev/null || git log --pretty=format:"%s%n%b")
echo "Commits since $LAST_TAG:"
echo "$COMMITS"

BUMP="patch"

if echo "$COMMITS" | grep -qE "^major:"; then
  BUMP="major" 
elif echo "$COMMITS" | grep -qE "^minor:"; then
  BUMP="minor"
elif echo "$COMMITS" | grep -qE "^patch:"; then
  BUMP="patch"
else
  echo "No version bump keyword found. Skipping version tagging."
  exit 0
fi  
  
VERSION=${LAST_TAG#v}
IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION"

case "$BUMP" in
  major)
    ((MAJOR+=1)); MINOR=0; PATCH=0 ;;
  minor)
    ((MINOR+=1)); PATCH=0 ;;
  patch)
    ((PATCH+=1)) ;;
esac

NEW_TAG="v$MAJOR.$MINOR.$PATCH"
echo "Creating tag $NEW_TAG"

if git rev-parse "$NEW_TAG" >/dev/null 2>&1; then
  echo "Tag $NEW_TAG already exists. Skipping tagging."
  exit 0
fi

git config user.name "SHAHANASSHA"
git config user.email "shashahanas5@gmail.com"

git tag -a "$NEW_TAG" -m "Release $NEW_TAG"
git push origin "$NEW_TAG"

echo "✅ Tagged and pushed $NEW_TAG successfully."
