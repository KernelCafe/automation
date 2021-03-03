#!/bin/sh
#
# Simple installer for Go - inspired by https://github.com/canha/golang-tools-install-script
set -eu -o pipefail

VERSION=${VERSION:-1.16}
GOOS=${GOOS:-""}
GOARCH=${GOARCH:-""}
GOROOT=${GOROOT:-"/usr/local/go"}

if [ -z "${GOROOT}" ]; then
  echo "GOROOT is empty"
  exit 2
fi

if [ -z "${GOOS}"]; then
    GOOS="$(uname -s | tr A-Z a-z)"
fi

if [ -z "${GOARCH}"]; then
    guess="$(uname -m | tr A-Z a-z)"
    case $guess in
    "x86_64")
        GOARCH=amd64
        ;;
    "aarch64")
        GOARCH=arm64
        ;;
    "armv6" | "armv7l")
        GOARCH=armv6l
        ;;
    "armv8")
        GOARCH=arm64
        ;;
    "i386")
        GOARCH=386
        ;;
    *)
        GOARCH=$guess
        ;;
    esac
fi

if [ ! -d "${GOROOT}" ]; then
  mkdir -p "${GOROOT}"
fi

target="${GOROOT}/bin/go"
if [ -x "${target}" ]; then
    inst="$(${target} version | cut -d" " -f3 | sed s/^go//g)"
    if [ "${inst}" = "${VERSION}" ]; then
        echo "go ${VERSION} is already installed"
        exit 0
    fi
    echo "Upgrading go from ${inst} to ${VERSION}"
fi

curl -L https://dl.google.com/go/go1.16.${GOOS}-${GOARCH}.tar.gz |
    tar --strip-components=1 -C "${GOROOT}" -xzf -

inst="$(${target} version | cut -d" " -f3 | sed s/^go//g)"
if [ "${inst}" != "${VERSION}" ]; then
    echo "go ${VERSION} installation failed?"
    exit 1
fi
