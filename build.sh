#!/bin/bash

### build backend
dir=`pwd`
version=0.0.1

xgo -out efront-v${version} --targets=windows/amd64,windows/386,darwin/amd64,linux/amd64 .
mv -f ${dir}/efront-v${version}-* bin
echo "assembly complete"
