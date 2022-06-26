#!/bin/bash

archs=(amd64 arm64 ppc64le ppc64 s390x)

for arch in ${archs[@]}
do
        env GOOS=linux GOARCH=${arch} go build -o dist/cubby_linux_${arch} `pwd`/cmd/client
done


marchs=(amd64 arm64)

for arch in ${marchs[@]}
do
        env GOOS=darwin GOARCH=${arch} go build -o dist/cubby_darwin_${arch} `pwd`/cmd/client
done

