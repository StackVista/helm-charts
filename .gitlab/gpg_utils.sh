#!/usr/bin/env bash
set -euo pipefail

# This script contains utilities for working with GPG keys and commit signing.

# Prints a log message to stderr with a standard prefix.
log() {
  echo "gpg_utils.sh: $1" >&2
}

# Ensure the script is being sourced rather than executed directly.
(return 0 2>/dev/null) && sourced=1 || sourced=0
if [ "$sourced" -eq 0 ]; then
  log "This script must be sourced, not executed directly."
  exit 1
fi

# Retrieves the GPG key ID associated with the STACKSTATE_SYSTEM_USER_EMAIL. To be called after import_gpg_key.
get_signing_key_id() {
  local signing_key_id
  signing_key_id=$(gpg --list-secret-keys --keyid-format LONG "$STACKSTATE_SYSTEM_USER_EMAIL" | grep 'sec ' | awk '{print $2}' | cut -d'/' -f2)
  log "Using GPG Key ID: $signing_key_id"
  if [ -z "$signing_key_id" ]; then
    log "Error: GPG Key ID not found for email $STACKSTATE_SYSTEM_USER_EMAIL"
    return 1
  fi
  echo "$signing_key_id"
}

# Checks if required tools (gpg, git, and gpg-agent) are available in the system.
validate_tools() {
  command -v gpg >/dev/null || (log "ERROR: 'gpg' command not found!" && exit 1)
  command -v git >/dev/null || (log "ERROR: 'git' command not found!" && exit 1)
  command -v gpg-agent >/dev/null || (log "ERROR: 'gpg-agent' command not found!" && exit 1)
}

# Imports the GPG key stored in STACKSTATE_SYSTEM_USER_PGP_KEY and verifies its import.
import_gpg_key() {
  log "Importing GPG key"
  # Directly read STACKSTATE_SYSTEM_USER_PGP_KEY content into gpg (file-type variable).
  gpg --batch --import "$STACKSTATE_SYSTEM_USER_PGP_KEY"
  gpg --list-secret-keys --keyid-format LONG "$STACKSTATE_SYSTEM_USER_EMAIL" || exit 1
}

# Checks if global git user.name, user.email, user.signingkey, and commit.gpgsign are configured.
check_git_configuration() {
  local user_name user_email user_signingkey commit_gpgsign signing_key_id

  user_name=$(git config --global user.name || echo "")
  user_email=$(git config --global user.email || echo "")
  user_signingkey=$(git config --global user.signingkey || echo "")
  commit_gpgsign=$(git config --global commit.gpgsign || echo "")

  if ! signing_key_id=$(get_signing_key_id); then
    log "Failed to retrieve the expected GPG Key ID."
    return 1
  fi

  if [[ "$user_name" == "$STACKSTATE_SYSTEM_USER_NAME" ]] &&
    [[ "$user_email" == "$STACKSTATE_SYSTEM_USER_EMAIL" ]] &&
    [[ "$user_signingkey" == "$signing_key_id" ]] &&
    [[ "$commit_gpgsign" == "true" ]]; then
    log "Git configuration is complete: '$user_name <$user_email>' with signing key '$user_signingkey' and commit.gpgsign set to '$commit_gpgsign'."
    return 0
  else
    log "Git configuration incomplete."
    return 1
  fi
}

# Configures the git user with name, email, and signing key, and starts the gpg-agent.
configure_git_user() {
  log "Configuring git user"

  git config --global user.name "$STACKSTATE_SYSTEM_USER_NAME"
  git config --global user.email "$STACKSTATE_SYSTEM_USER_EMAIL"
  log "Git user configured as '$STACKSTATE_SYSTEM_USER_NAME <$STACKSTATE_SYSTEM_USER_EMAIL>'."

  validate_tools
  import_gpg_key

  local signing_key_id
  if ! signing_key_id=$(get_signing_key_id); then
    log "Error: Unable to retrieve GPG Key ID. Aborting configuration."
    exit 1
  fi

  git config --global user.signingkey "$signing_key_id"
  git config --global commit.gpgsign true

  eval "$(gpg-agent --daemon)"
  echo "$STACKSTATE_SYSTEM_USER_PGP_PASS_PHRASE" | gpg --batch --trust-model always --yes --pinentry-mode loopback --passphrase-fd 0 --status-fd 1 --sign --detach-sign --output /dev/null /dev/null
}
