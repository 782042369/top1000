#!/bin/bash

# 默认初始版本号
DEFAULT_VERSION="1.7"

# 使用函数处理版本号增加
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

# 使用函数更新文件中的版本号
update_version_in_file() {
    local file=$1
    local version=$2
    local pattern=$3
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i "" "s#$pattern#$version#" "$file"
    else
        sed -i "s#$pattern#$version#" "$file"
    fi
}

# 主逻辑
main() {
    # 读取当前版本号或使用默认版本号
    local current_version=$(grep '^VERSION=' $0 | cut -d'=' -f2 | tr -d '"')
    local version=${1:-$(increment_version ${current_version:-$DEFAULT_VERSION})}

    # 更新脚本中的版本号
    update_version_in_file "$0" "VERSION=\"$version\"" "^VERSION=\"[0-9]*\.[0-9]*\""

    # 更新 docker-compose.yaml 文件中的版本号
    update_version_in_file "docker-compose.yaml" "image: 782042369/top1000-iyuu:v$version" "image: 782042369\/top1000-iyuu:v[0-9]*\.[0-9]*"

    echo "Building version $version"

    # 构建过程
    cd web && pnpm i && pnpm build
    cd ../service && docker buildx build --platform linux/amd64,linux/arm64 -t 782042369/top1000-iyuu:$version . --push

    # Git 提交
    cd ../
    git add .
   git commit -m "feat: docker build v.$VERSION"
    git push

    echo "Updated script and docker-compose to version $version and pushed to Git."
}

# 调用主函数
main "$@"
