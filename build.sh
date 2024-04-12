#!/bin/bash

# 初始版本号
VERSION="1.6"

# 自动增加版本号的函数
increment_version() {
    local IFS=.
    local num=($1)
    if (( num[1] == 9 )); then
        ((num[0]++))
        num[1]=0
    else
        ((num[1]++))
    fi
    echo "${num[0]}.${num[1]}"
}

# 读取当前版本号
current_version=$(grep '^VERSION=' $0 | cut -d'=' -f2 | tr -d '"')

# 检查是否传入了版本号或使用自动增加的版本号
if [ "$#" -ne 1 ]; then
    if [ -z "$current_version" ]; then
        echo "No version specified and no version found in script."
        exit 1
    fi
    VERSION=$(increment_version $current_version)
else
    VERSION=$1
fi

# 更新脚本中的版本号
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS requires an empty extension with -i for in-place editing
    sed -i "" "s/^VERSION=\"[0-9]*\.[0-9]*\"/VERSION=\"$VERSION\"/" $0
else
    # Linux and others
    sed -i "s/^VERSION=\"[0-9]*\.[0-9]*\"/VERSION=\"$VERSION\"/" $0
fi

echo "Building version $VERSION"

# 构建过程
cd web && pnpm i && pnpm build
cd ../service && docker buildx build --platform linux/amd64,linux/arm64 -t 782042369/top1000-iyuu:v.$VERSION . --push

# Git 提交
git add $0
git commit -m "feat: docker build v.$VERSION"
git push

echo "Updated script to version $VERSION and pushed to Git."
