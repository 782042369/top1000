#!/bin/bash

# 检查是否传入了版本号
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <version>"
    exit 1
fi

VERSION=$1

cd web && pnpm i && pnpm build
cd ../service && docker buildx build --platform linux/amd64,linux/arm64 -t 782042369/top1000-iyuu:$VERSION . --push
