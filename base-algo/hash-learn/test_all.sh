#!/bin/bash

echo "========================================="
echo "测试所有哈希学习示例"
echo "========================================="
echo ""

echo "1. 测试基础哈希函数..."
echo "----------------------------------------"
go run examples/basic_hash/main.go > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ 基础哈希函数示例运行成功"
else
    echo "✗ 基础哈希函数示例运行失败"
    exit 1
fi
echo ""

echo "2. 测试密码哈希..."
echo "----------------------------------------"
go run examples/password_hash/main.go > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ 密码哈希示例运行成功"
else
    echo "✗ 密码哈希示例运行失败"
    exit 1
fi
echo ""

echo "3. 测试一致性哈希..."
echo "----------------------------------------"
go run examples/consistent_hash/main.go > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ 一致性哈希示例运行成功"
else
    echo "✗ 一致性哈希示例运行失败"
    exit 1
fi
echo ""

echo "4. 测试MD5实现..."
echo "----------------------------------------"
go run md5/md5.go > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ MD5实现运行成功"
else
    echo "✗ MD5实现运行失败"
    exit 1
fi
echo ""

echo "========================================="
echo "所有测试通过！✓"
echo "========================================="
