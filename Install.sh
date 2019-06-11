#!/usr/bin/env bash

PARAMS=""
SKIP_GET = 0
GOPATH="`cd ~ && pwd`/go"
GITHUB_USER='portofportland'
PROVIDER_NAME='terraform-provider-windns'

while (( "$#" )); do
  case "$1" in
    -g|--gopath)
      GOPATH=$2
      shift 2
      ;;
    -u|--githubUser)
      GITHUB_USER=$2
      shift 2
      ;;
    -p|--providerName)
      PROVIDER_NAME=$2
      shift 2
      ;;
    --skipGet)
      SKIP_GET=1
      shift 2
      ;;
    --) # end argument parsing
      shift
      break
      ;;
    -*|--*=) # unsupported flags
      echo "Error: Unsupported flag $1" >&2
      exit 1
      ;;
    *) # preserve positional arguments
      PARAMS="$PARAMS $1"
      shift
      ;;
  esac
done
# set positional arguments in their proper place
eval set -- "$PARAMS"

PROVIDER_REPO="github.com/${GITHUB_USER}/${PROVIDER_NAME}"

mkdir -p ${GOPATH}
go get ${PROVIDER_REPO}

BIN_PATH="${GOPATH}/bin/${PROVIDER_NAME}"

cd ${GOPATH}/src/${PROVIDER_REPO}
rm ${BIN_PATH}
go build
go install

TERRAFORM_PLUGINS_DIR="`cd ~ && pwd`/.terraform.d/plugins"
mkdir -p ${TERRAFORM_PLUGINS_DIR}
PROVIDER_PATH="${TERRAFORM_PLUGINS_DIR}/${PROVIDER_NAME}"

if [ -f "${BIN_PATH}" ]; then
  cp "${BIN_PATH}" ${PROVIDER_PATH}
  echo "Copy Successful.  ${PROVIDER_PATH}"
else
  echo 'Build Failed, Copy Aborted.'
  exit 1
fi

exit 0