#!/usr/bin/bash

cd go_tooling
GOOS=js GOARCH=wasm go build -o ../lib/go_helpercode.wasm steply
cd ../

sudo cp -r lib index.html /var/www/html/steply