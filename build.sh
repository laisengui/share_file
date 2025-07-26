#!/bin/sh
##################################
# 生成各个平台下的可执行程序 golang一键打包 macos, linux, windows 应用程序
# 使用方法: sh build.sh [-n appname]
# 也可忽略 -n 参数 sh build.sh  默认名称为 myapp
#
# 如: sh build.sh -n helloworld 将自动在target目录下生成以下3个可执行文件
# helloworld-darwin-amd64.bin  helloworld-linux-amd64.bin helloworld-windows-amd64.exe
# 如指定版本: sh build.sh -v 0.0.2 将自动在target目录下生成以下2个可执行文件
#
# Author: tekintian@gmail.com
##################################

APPNAME="share_file"
# 通用变量
export CGO_ENABLED=0 # 关闭CGO
export GOARCH=amd64  #CPU架构
# 设置darwin
export GOOS=darwin
go build -ldflags "-s -w" -o target/${APPNAME}-darwin-amd64.bin
echo "Macos可执行程序 ${APPNAME}-darwin-amd64.bin 打包成功!"
# 设置linux
export GOOS=linux
go build -ldflags "-s -w" -o target/${APPNAME}-linux-amd64.bin
echo "linux可执行程序 ${APPNAME}-linux-amd64.bin 打包成功!"

# 设置windows
export GOOS=windows
go build -ldflags "-s -w" -o target/${APPNAME}-windows-amd64.exe
echo "Windows可执行程序 ${APPNAME}-windows-amd64.exe 打包成功!"

export GOOS=linux
export GOARCH=arm64  #CPU架构
go build -ldflags "-s -w" -o target/${APPNAME}-linux-arm64.bin
echo "linux可执行程序 ${APPNAME}-linux-arm64.bin 打包成功!"

export GOOS=darwin
go build -ldflags "-s -w" -o target/${APPNAME}-darwin-arm64.bin
echo "Macos可执行程序 ${APPNAME}-darwin-arm64.bin 打包成功!"
