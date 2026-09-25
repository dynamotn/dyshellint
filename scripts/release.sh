#!/usr/bin/env bash
# @file release.sh
# @brief Cut a release of dyshellint from the current repository
# @description Decides the next version from the Conventional Commits made
#   since the last tag, runs the gates, turns the `## [Unreleased]` section of
#   CHANGELOG.md into the entry for that version, then commits, tags and pushes.
#   The build and the release itself are left to goreleaser, which the `release`
#   workflow runs when the tag lands, so nothing here needs a forge token.
#
#   Every state-changing step goes through `dybatpho::dry_run`, so `--dry-run`
#   shows the whole plan without touching the repository or the remote.
#
#   Usage examples:
#     bash ./scripts/release.sh --dry-run
#     bash ./scripts/release.sh --bump minor
#     bash ./scripts/release.sh --version 1.4.0 --yes
DYBATPHO_PATH="${DYBATPHO_DIR:-${HOME}/Dotfiles/scripts/lib/dybatpho}"
if [[ ! -r "${DYBATPHO_PATH}/init.sh" ]]; then
  printf 'dybatpho not found at %s; set DYBATPHO_DIR to its checkout\n' \
    "${DYBATPHO_PATH}" >&2
  exit 1
fi
# shellcheck source=/dev/null
. "${DYBATPHO_PATH}/init.sh" --modules release safety cli
dybatpho::register_common_handlers

readonly REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
readonly CHANGELOG="${REPO_DIR}/CHANGELOG.md"
# The pattern the tags of this repository follow, overridable for a fork that
# numbers its tags differently.
readonly TAG_PATTERN="${DYBATPHO_RELEASE_TAG_PATTERN:-v*}"

# REMOTE is declared by `_spec_main` and filled in by
# `dybatpho::generate_from_spec`, which ShellCheck cannot see through.
# shellcheck disable=SC2154
#######################################
# @description Refuse to release from a tree that does not match what the tag
#   will claim
# @noargs
# @exitcode 0 If the repository is ready to release
#######################################
function _preflight {
  dybatpho::require git
  dybatpho::require go
  dybatpho::require make

  local branch default_branch
  branch="$(dybatpho::git_branch "${REPO_DIR}")"
  default_branch="$(dybatpho::git_default_branch "${REPO_DIR}")"
  if [[ "${branch}" != "${default_branch}" ]]; then
    dybatpho::confirm \
      "On branch ${branch}, not ${default_branch}. Release from it anyway?" \
      || dybatpho::die "Switch to ${default_branch} before releasing"
  fi

  dybatpho::git_is_clean "${REPO_DIR}" \
    || dybatpho::die "Worktree has changes; commit or stash them first"

  dybatpho::git_has_remote "${REMOTE}" "${REPO_DIR}" \
    || dybatpho::die "No remote named ${REMOTE}"
}

#######################################
# @description Decide which version to release
# @description Sets a variable rather than printing, so that a rejected version
#   stops the script itself instead of only the command substitution that would
#   read it.
# @noargs
# @set RELEASE_VERSION Version to release, without a leading `v`
# @exitcode 0 If a version could be decided
#######################################
function _resolve_version {
  if [[ -n "${VERSION:-}" ]]; then
    RELEASE_VERSION="${VERSION#v}"
    dybatpho::semver_valid "${RELEASE_VERSION}" \
      || dybatpho::die "Not a valid semantic version: ${VERSION}"
  elif [[ -n "${BUMP:-}" ]]; then
    [[ -n "${PREVIOUS_TAG}" ]] \
      || dybatpho::die "No ${TAG_PATTERN} tag to bump from; pass --version"
    RELEASE_VERSION="$(dybatpho::semver_bump "${PREVIOUS_TAG#v}" "${BUMP}")"
  else
    # No commit that calls for a release leaves this empty, which is a stop,
    # not a failure: there is simply nothing to ship.
    RELEASE_VERSION="$(dybatpho::release_next_version "${REPO_DIR}" || true)"
    [[ -n "${RELEASE_VERSION}" ]] \
      || dybatpho::die "No feat/fix/perf commit since ${PREVIOUS_TAG:-the first commit}; nothing to release"
  fi
}

#######################################
# @description Stop when the tag of a version already exists
# @arg $1 string Version being released, without a leading `v`
# @exitcode 0 If the tag is free
#######################################
function _check_tag_free {
  local version
  dybatpho::expect_args version -- "$@"

  local tag="v${version}"
  if git -C "${REPO_DIR}" rev-parse -q --verify "refs/tags/${tag}" > /dev/null; then
    dybatpho::die "Tag ${tag} already exists"
  fi
}

#######################################
# @description Print the body of the `## [Unreleased]` section, without its
#   heading
# @description Entries are written by hand as the change is made, so this is the
#   list that knows what a release actually contains; the one derived from
#   commit subjects only knows what they were called.
# @noargs
# @stdout The entries, or nothing when the section is empty or absent
# @exitcode 0 Always
#######################################
function _unreleased_body {
  [[ -f "${CHANGELOG}" ]] || return 0
  # shellcheck disable=SC2312
  sed -n '/^## \[Unreleased\]/,/^## \[/{ /^## \[/d; p; }' "${CHANGELOG}" \
    | sed -e '/./,$!d' \
    | awk 'BEGIN { blank = 0 }
           /^$/ { blank++; next }
           { while (blank-- > 0) print ""; blank = 0; print }'
}

#######################################
# @description Turn the `## [Unreleased]` section into the entry of a version
# @description The new section goes directly under the preamble, so the file
#   stays newest-first the way Keep a Changelog describes, and the hand-written
#   entries move with it: leaving them above the version they shipped in would
#   strand them there for good.
# @arg $1 string Version being released, without a leading `v`
# @exitcode 0 If the changelog was rewritten
#######################################
function _write_changelog {
  local version
  dybatpho::expect_args version -- "$@"

  local entry unreleased
  entry="$(dybatpho::release_changelog \
    "${REPO_DIR}" "${PREVIOUS_TAG}" HEAD "${version}")"
  unreleased="$(_unreleased_body)"
  if [[ -n "${unreleased}" ]]; then
    dybatpho::info "Releasing the hand-written Unreleased entries"
    entry="$(printf '## [%s]\n\n%s' "${version}" "${unreleased}")"
  fi

  # Everything above the first version heading is the preamble, and everything
  # from it on is history, minus the Unreleased section that just became one.
  local preamble history
  preamble="$(sed '/^## \[/,$d' "${CHANGELOG}")"
  # shellcheck disable=SC2312
  history="$(sed -n '/^## \[/,$p' "${CHANGELOG}" \
    | sed '/^## \[Unreleased\]/,/^## \[/{ /^## \[Unreleased\]/d; /^## \[/!d; }')"

  local temp_dir temp_file
  dybatpho::create_temp_dir temp_dir
  temp_file="${temp_dir}/CHANGELOG.md"
  {
    if [[ -n "${preamble}" ]]; then
      printf '%s\n\n' "${preamble}"
    fi
    # An empty Unreleased section is left behind, so the next change has a
    # place to be written down the moment it is made.
    printf '## [Unreleased]\n\n'
    printf '%s\n' "${entry}"
    if [[ -n "${history}" ]]; then
      printf '\n%s\n' "${history}"
    fi
  } > "${temp_file}"

  dybatpho::header "CHANGELOG ENTRY"
  dybatpho::print "${entry}"

  dybatpho::dry_run cp "${temp_file}" "${CHANGELOG}"
}

#######################################
# @description Run the same gates as CI before anything is tagged
# @noargs
# @exitcode 0 If every gate passes
#######################################
function _run_gates {
  if dybatpho::is true "${SKIP_CHECKS:-false}"; then
    dybatpho::warn "Skipping lint and tests"
    return 0
  fi
  dybatpho::info "Running lint"
  dybatpho::dry_run make -C "${REPO_DIR}" lint
  dybatpho::info "Running tests"
  dybatpho::dry_run make -C "${REPO_DIR}" test
}

#######################################
# @description Commit the changelog, tag the release and push it, which is what
#   starts the workflow
# @arg $1 string Version being released, without a leading `v`
# @exitcode 0 If the tag is created, and pushed unless --skip-push
#######################################
function _publish {
  local version
  dybatpho::expect_args version -- "$@"

  local tag="v${version}"
  local branch
  branch="$(dybatpho::git_branch "${REPO_DIR}")"

  dybatpho::dry_run git -C "${REPO_DIR}" add CHANGELOG.md
  dybatpho::dry_run git -C "${REPO_DIR}" commit -m "chore(release): ${tag}"
  dybatpho::dry_run git -C "${REPO_DIR}" tag -a "${tag}" -m "Release ${tag}"

  if dybatpho::is true "${SKIP_PUSH:-false}"; then
    dybatpho::warn "Not pushing; run: git push ${REMOTE} ${branch} && git push ${REMOTE} ${tag}"
    return 0
  fi
  dybatpho::dry_run git -C "${REPO_DIR}" push "${REMOTE}" "${branch}"
  dybatpho::dry_run git -C "${REPO_DIR}" push "${REMOTE}" "${tag}"
  dybatpho::info "Pushed ${tag}; the release workflow builds and publishes it"
}

# The options below are declared by `_spec_main` and filled in by
# `dybatpho::generate_from_spec`, which ShellCheck cannot see through.
# shellcheck disable=SC2154
#######################################
# @description Decide the version, run the gates, write the changelog, then
#   commit, tag and push
# @noargs
# @set PREVIOUS_TAG Last tag matching the pattern, empty on the first release
# @exitcode 0 If the release was cut
#######################################
function _release {
  if dybatpho::is true "${DRY_RUN:-false}"; then
    dybatpho::warn "Dry run: nothing will be changed"
  fi

  _preflight

  PREVIOUS_TAG="$(dybatpho::git_latest_tag "${REPO_DIR}" "${TAG_PATTERN}" || true)"
  dybatpho::info "Previous release: ${PREVIOUS_TAG:-none}"

  _resolve_version
  local version="${RELEASE_VERSION}"
  _check_tag_free "${version}"
  dybatpho::info "Releasing v${version}"

  _run_gates
  _write_changelog "${version}"

  dybatpho::confirm "Commit, tag and push v${version}?" yes \
    || dybatpho::die "Release aborted"

  _publish "${version}"
  dybatpho::success "Released v${version}"
}

#######################################
# @description Spec of release.sh
# @noargs
#######################################
function _spec_main {
  dybatpho::opts::setup \
    "Cut a release of dyshellint: version, changelog, tag and push" \
    RELEASE_ARGS action:"_release"

  dybatpho::opts::param "Version to release, instead of deriving one" VERSION -v --version
  dybatpho::opts::param "Bump the last tag by this part (major|minor|patch)" BUMP -b --bump
  dybatpho::opts::param "Remote to push to" REMOTE -r --remote init:="origin"
  dybatpho::opts::flag "Show every step without changing anything" DRY_RUN -n --dry-run
  dybatpho::opts::flag "Skip lint and tests" SKIP_CHECKS --skip-checks
  # The switch is `--skip-push`, not `--no-push`: dybatpho reads a leading
  # `--no-` as the negation of a flag, so `--no-push` would never set this one.
  dybatpho::opts::flag "Tag, but do not push" SKIP_PUSH --skip-push
  dybatpho::opts::flag "Answer yes to every prompt" DYBATPHO_FORCE -y --yes

  dybatpho::opts::disp "Show help" -h --help action:"dybatpho::generate_help _spec_main"
}

dybatpho::generate_from_spec _spec_main "$@"
