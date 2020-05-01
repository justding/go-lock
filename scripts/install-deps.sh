#!/usr/bin/env bash

if ! [[ -x "$(command -v protoc)" ]]; then
    echo 'installing protoc...'
    if [[ "$OSTYPE" == "darwin"* ]]; then
        brew install protobuf
    else
        PROTOC_ZIP=protoc-3.7.1-linux-x86_64.zip
        curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v3.7.1/$PROTOC_ZIP
        sudo unzip -o $PROTOC_ZIP -d /usr/local bin/protoc
        sudo unzip -o $PROTOC_ZIP -d /usr/local 'include/*'
        rm -f $PROTOC_ZIP
    fi
else
    echo 'protoc already installed!'
fi

if ! [[ -x "$(command -v protoc-gen-go)" ]]; then
    echo 'installing protoc-gen-go utility...'
    go install google.golang.org/protobuf/cmd/protoc-gen-go
else
    echo 'protoc-gen-go already installed!'
fi
