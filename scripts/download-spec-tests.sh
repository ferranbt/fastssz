#!/bin/bash

VERSION=$1
REPO_NAME=eth2.0-spec-tests

# Remove dir if it already exists
rm -rf $REPO_NAME
mkdir $REPO_NAME

function download {
    OUTPUT=$1.tar.gz
    DOWNLOAD_URL=https://github.com/ethereum/$REPO_NAME/releases/download/$VERSION/$OUTPUT
    echo $DOWNLOAD_URL
    wget $DOWNLOAD_URL -O $OUTPUT
    tar -xzf $OUTPUT -C $REPO_NAME
    rm $OUTPUT
}

download "mainnet"
download "minimal"
