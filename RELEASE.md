# Release procedure for Blubber

Run `scripts/release.sh` which will:

 1. Increment value in `VERSION` (minor by default; pass `-p` to do a
    patch release).
 2. Generate `CHANGELOG.md` using `git chglog`.
 3. Create a commit for the new version and change log.
 4. Create a signed version tag.
 5. Push the new commit and version tag.
