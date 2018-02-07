#!/bin/bash

### build backend
dir=`pwd`

xgo -out efront --targets=windows/amd64,darwin/amd64,linux/amd64 .
mv -f ${dir}/efront-* bin
echo "assembly complete"
