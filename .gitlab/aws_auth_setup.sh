#!/usr/bin/env bash

# Fixing up this bug AWS made for service accounts with non-root users:
# https://github.com/aws/amazon-eks-pod-identity-webhook/issues/8

new_token_file="/$CI_PROJECT_DIR/aws_token"
cp "${AWS_WEB_IDENTITY_TOKEN_FILE}" "${new_token_file}"
chmod a+r "${new_token_file}"

# Ignore the settings from the environment such that the config file is used instead
export AWS_WEB_IDENTITY_TOKEN_FILE="${new_token_file}"
