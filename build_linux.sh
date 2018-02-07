#!/bin/bash

### build backend
dir=`pwd`
version=0.0.1

env GOOS=linux go build -o bin/efront main.go
echo "assembly complete"
