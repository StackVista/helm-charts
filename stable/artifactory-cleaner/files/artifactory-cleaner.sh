#!/bin/bash

set -euo pipefail

: "${ARTIFACTORY_URL:?ARTIFACTORY_URL is not set}"
: "${ARTIFACTORY_USER:?ARTIFACTORY_USER is not set}"
: "${ARTIFACTORY_PASSWORD:?ARTIFACTORY_PASSWORD is not set}"
: "${RETENTION_MANIFEST:?RETENTION_MANIFEST is not set}"

DIR="${BASH_SOURCE%/*}"
if [[ ! -d "${DIR}" ]]; then DIR="$PWD"; fi
if [[ "${DIR}" = "." ]]; then DIR="$PWD"; fi

APPLY=${1:-no}

function echoerr() {
  echo "$@" 1>&2
}

function create_aql_query() {
  query_type=$1
  query_arg=$2
  repo=$3
  repo_path=$4
  exclude=$5
  name_filter=$6

  case "${query_type}" in
    keep-last)
      offset="${query_arg}"
      until_filter=""
      ;;
    time-based)
      offset=""
      until_filter="${query_arg}"
      ;;
  esac

  if [ "${repo}" = "null" ]; then
    echoerr "repo must be specified. exiting..."
    exit 1
  fi

  repo_path_filter=""
  if [ "${repo_path}" != "null" ]; then
    repo_path_filter="\"path\":{\"\$match\" : \"${repo_path}/*\"},"
  fi

  exclude_path_filter=""
  if [ "${exclude}" != "null" ]; then
    exclude_path_filter="\"path\":{\"\$nmatch\" : \"${exclude}\"},"
  fi

  name_filter_filter=""
  if [ "${name_filter}" != "null" ]; then
    name_filter_filter="\"name\":{\"\$match\":\"${name_filter}\"},"
  fi

  cat <<EOF
items
  .find({
    "repo":{"\$eq":"${repo}"},
    ${repo_path_filter}
    ${exclude_path_filter}
    ${until_filter}
    ${name_filter_filter}
    "type":"file"
  })
  .include("name", "repo", "path", "created")
  ${offset}
EOF
}

function get_artifacts() {
  query=$1
  query_result=$(echo "${query}" | /usr/bin/curl -s -u "${ARTIFACTORY_USER}:${ARTIFACTORY_PASSWORD}" -X POST -k "${ARTIFACTORY_URL}/api/search/aql" -H "Content-Type: text/plain" --data @-)
  curl_exit_code=$?
  if [ "${curl_exit_code}" != "0" ]; then
    echoerr "curl command (-X POST -k ${ARTIFACTORY_URL}/api/search/aql) failed with the exit code: ${curl_exit_code}"
    exit 1
  fi
  echo "${query_result}" | jq '.results | .[]'
}

function delete_artifacts() {
  artifacts=$1
  apply=$2

  for item in $(echo "${artifacts}" | jq -r '[.path] | unique | .[]' | sort | uniq)
  do
    echo "Deleting ${item} : ${ARTIFACTORY_URL}/${repo}/${item}"
    if [ "${apply}" = "--apply" ]; then
      /usr/bin/curl -s -u "${ARTIFACTORY_USER}:${ARTIFACTORY_PASSWORD}" -X DELETE -k "${ARTIFACTORY_URL}/${repo}/${item}"
      curl_exit_code=$?
      if [ "${curl_exit_code}" != "0" ]; then
        echoerr "curl command (-X DELETE -k ${ARTIFACTORY_URL}/${repo}/${item}) failed with the exit code: ${curl_exit_code}"
        exit 1
      fi
      sleep 1
    fi
  done
}

jq -c '.[]' "${RETENTION_MANIFEST}" | while read -r policy
do
  query_type=$(echo "${policy}" | jq -r .type)

  case "${query_type}" in
    keep-last)
      keep=$(echo "${policy}" | jq -r .artifactsToKeep)
      query_arg=".sort({\"\$desc\": [\"created\"]}).offset(${keep})"
      ;;
    time-based)
      age=$(echo "${policy}" | jq -r .age)
      until=$(date -d "${age}" +%Y-%m-%d)
      query_arg="\"created\":{\"\$lt\":\"${until}\"},"
      ;;
    *)
      echo "unknown query type (${query_type}). must be \"keep-last\" or \"keep-last\". exiting..."
      exit 1
      ;;
  esac

  repo=$(echo "${policy}" | jq -r .repo)
  repo_path=$(echo "${policy}" | jq -r .path)
  name_filter=$(echo "${policy}" | jq -r .nameFilter)
  exclude=$(echo "${policy}" | jq -r .exclude)

  aql_query=$(create_aql_query "${query_type}" "${query_arg}" "${repo}" "${repo_path}" "${exclude}" "${name_filter}")
  artifacts=$(get_artifacts "${aql_query}")
  delete_artifacts "${artifacts}" "${APPLY}"

done
