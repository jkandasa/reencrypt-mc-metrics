#!/bin/bash

# this script used to generate binary files
BUILD_DIR=builds

# platforms to build
PLATFORMS=("linux/arm" "linux/arm64" "linux/386" "linux/amd64" "linux/ppc64" "linux/ppc64le" "linux/s390x" "darwin/amd64" "windows/386" "windows/amd64")

rm ${BUILD_DIR} -rf

# compile
for platform in "${PLATFORMS[@]}"
do
  platform_raw=(${platform//\// })
  GOOS=${platform_raw[0]}
  GOARCH=${platform_raw[1]}
  package_name="repack-mc-metrics-${GOOS}-${GOARCH}"

  FILE_EXTENSION=""
  if [ $GOOS = "windows" ]; then
    FILE_EXTENSION='.exe'
  fi

  env GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 go build -o ${BUILD_DIR}/${package_name}${FILE_EXTENSION} -ldflags "-s -w" main.go
  if [ $? -ne 0 ]; then
    echo 'an error has occurred. aborting the build process'
    exit 1
  fi


done