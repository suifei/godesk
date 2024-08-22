#!/bin/bash

# 设置输出文件名
output_file="golang-project-godesk-codebase.txt"

# 清空输出文件（如果已存在）
> "$output_file"

# 查找并处理所有的 .go, .proto, .mod, 和 .sh 文件
find . -type f \( -name "*.go" -o -name "*.proto" -o -name "*.mod" -o -name "*.sh" \) | sort | while read -r file
do
    # 获取相对路径
    relative_path="${file#./}"
    
    # 将文件路径写入输出文件
    echo "File: $relative_path" >> "$output_file"
    echo "----------------------------------------" >> "$output_file"
    
    # 将文件内容写入输出文件
    cat "$file" >> "$output_file"
    
    # 添加分隔符
    echo -e "\n\n========================================\n\n" >> "$output_file"
done

echo "所有文件已合并到 $output_file"