#!/bin/bash

# 确保脚本在发生错误时退出
set -e

# 定义版本号（你可以根据需要修改）
VERSION="0.1.0"

# 定义输出目录
OUTPUT_DIR="build"

# 定义目标操作系统和架构
TARGETS=(
    "windows/amd64"
    #"linux/amd64"
    #"darwin/amd64"
)

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

# 编译函数
build() {
    local module=$1
    local os=$2
    local arch=$3
    local output_name="${module}_${os}_${arch}"
    if [ "$os" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    echo "Building $module for $os/$arch..."
    GOOS=$os GOARCH=$arch go build -o "$OUTPUT_DIR/$output_name" "./cmd/$module"
}

# 为每个目标构建所有模块
for target in "${TARGETS[@]}"; do
    IFS='/' read -r -a parts <<< "$target"
    os="${parts[0]}"
    arch="${parts[1]}"
    
    build "server" "$os" "$arch"
    build "client" "$os" "$arch"
    build "relay" "$os" "$arch"
done

# 复制配置文件
echo "Copying configuration files..."
mkdir -p "$OUTPUT_DIR/configs"
cp configs/*.yaml "$OUTPUT_DIR/configs/"

# 创建版本文件
echo "$VERSION" > "$OUTPUT_DIR/version.txt"

echo "Build complete. Output is in the $OUTPUT_DIR directory."

# 显示构建输出的结构
echo "Build output structure:"
if command -v tree &> /dev/null; then
    tree "$OUTPUT_DIR"
else
    find "$OUTPUT_DIR" | sed -e "s/[^-][^\/]*\// |/g" -e "s/|\([^ ]\)/|-\1/"
fi