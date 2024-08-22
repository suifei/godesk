#!/bin/bash

# 检查 protoc 是否已安装
if ! command -v protoc &> /dev/null
then
    echo "protoc could not be found. Please install Protocol Buffers compiler."
    echo "Visit https://github.com/protocolbuffers/protobuf for installation instructions."
    exit 1
fi

# 检查 protoc-gen-go 是否已安装
if ! command -v protoc-gen-go &> /dev/null
then
    echo "protoc-gen-go could not be found. Installing it now..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    
    # 将 GOPATH/bin 添加到 PATH
    export PATH="$PATH:$(go env GOPATH)/bin"
fi

# 进入项目根目录
cd "$(dirname "$0")/.."

# 确保 internal/protocol 目录存在
mkdir -p internal/protocol

# 生成 Go 代码，并指定输出目录为 internal/protocol
protoc --go_out=. --go_opt=module=github.com/suifei/godesk \
    --go_opt=Mproto/screen.proto=github.com/suifei/godesk/internal/protocol \
    --go_opt=Mproto/control.proto=github.com/suifei/godesk/internal/protocol \
    --go_opt=Mproto/auth.proto=github.com/suifei/godesk/internal/protocol \
    --go_opt=Mproto/filetransfer.proto=github.com/suifei/godesk/internal/protocol \
    --go_opt=Mproto/relay.proto=github.com/suifei/godesk/internal/protocol \
    --go_opt=Mproto/message.proto=github.com/suifei/godesk/internal/protocol \
    proto/screen.proto \
    proto/control.proto \
    proto/auth.proto \
    proto/filetransfer.proto \
    proto/relay.proto \
    proto/message.proto

if [ $? -eq 0 ]; then
    echo "Protocol Buffers code generation completed successfully."
    echo "Generated files are in the internal/protocol directory."
else
    echo "Error occurred during Protocol Buffers code generation."
    exit 1
fi

# 移除可能在 proto 目录中生成的 .pb.go 文件
rm -f proto/*.pb.go