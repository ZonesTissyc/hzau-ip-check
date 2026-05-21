#!/bin/bash

# 定义要检查的文件名（包含你提到的错别字版本和官方标准版本）
FILE="qqwry.dat"

# 最新版的下载地址
URL="https://github.com/metowolf/qqwry.dat/releases/latest/download/qqwry.dat"

if [ -f "$FILE1" ]; then
    echo "检测到当前目录已存在 IP 数据库文件，跳过下载。"
else
    echo "未检测到相关文件，开始下载最新版 qqwry.dat..."
    wget "$URL"
fi