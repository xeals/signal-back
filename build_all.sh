#!/usr/bin/env bash

set -e

BUILD_OSS="linux windows darwin"
BUILD_ARCHS="386 amd64"
BUILD_ROOT="$(git rev-parse --show-toplevel)"
BUILD_VER="$(git describe --tags --dirty)"

if [[ "$(pwd)" != "${BUILD_ROOT}" ]] ; then
    echo "must be in build root '${BUILD_ROOT}'"
    exit 1
fi

if [[ ! -d "${BUILD_ROOT}/release" ]] ; then
    mkdir -p "${BUILD_ROOT}/release"
fi

for OS in ${BUILD_OSS[@]} ; do
    for ARCH in ${BUILD_ARCHS[@]} ; do
        NAME="signal-back_${OS}_${ARCH}"
        echo "Building for ${OS}/${ARCH}"

        # .exe extension
        if [[ ${OS} == "windows" ]] ; then
            NAME="${NAME}.exe"
        fi

        GOOS=$OS GOARCH=$ARCH go build -ldflags "-X main.version=${BUILD_VER}" \
        -o "${BUILD_ROOT}/release/${NAME}" .
        shasum -a 256 "${BUILD_ROOT}/release/${NAME}" > "${BUILD_ROOT}/release/${NAME}.sha256"
    done
done
