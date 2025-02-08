#!/bin/bash
###
 # @Author: yanghongxuan
 # @Date: 2025-02-08 13:41:26
 # @Description:
 # @LastEditTime: 2025-02-08 14:21:52
 # @LastEditors: yanghongxuan
###

# 默认初始版本号
DEFAULT_VERSION="2.3"

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
    update_version_in_file "$0" "DEFAULT_VERSION=\"$version\"" "^DEFAULT_VERSION=\"[0-9]*\.[0-9]*\""


    echo "Building version $version"

    # 构建过程
    cd web && pnpm i && pnpm build
    cd ../service && docker build -t 782042369/top1000-iyuu:v.$version . --push

    # Git 提交
    cd ../
    git add .
    git commit -m "feat: docker build v.$DEFAULT_VERSION"
    git push

    echo "Updated script to version $version and pushed to Git."
}

# 调用主函数
main "$@"
