package main

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "mySecretPassword123"

	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println("bcrypt 密码哈希演示")
	fmt.Println("=" + string(make([]byte, 70)) + "=")

	// 测试不同的cost值
	costs := []int{10, 12, 14}

	for _, cost := range costs {
		start := time.Now()
		hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("Error with cost %d: %v\n", cost, err)
			continue
		}

		fmt.Printf("\nCost %d (耗时: %v):\n", cost, elapsed)
		fmt.Printf("Hash: %s\n", string(hash))

		// 验证密码
		err = bcrypt.CompareHashAndPassword(hash, []byte(password))
		if err == nil {
			fmt.Println("密码验证: ✓ 成功")
		} else {
			fmt.Println("密码验证: ✗ 失败")
		}

		// 测试错误密码
		err = bcrypt.CompareHashAndPassword(hash, []byte("wrongPassword"))
		if err != nil {
			fmt.Println("错误密码验证: ✓ 正确拒绝")
		}
	}

	// 演示相同密码的不同哈希（因为盐值不同）
	fmt.Println("\n" + "=" + string(make([]byte, 70)) + "=")
	fmt.Println("相同密码的多次哈希（盐值不同）:")
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	for i := 0; i < 3; i++ {
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
		fmt.Printf("%d: %s\n", i+1, string(hash))
	}

	// 演示bcrypt的安全性特点
	fmt.Println("\n" + "=" + string(make([]byte, 70)) + "=")
	fmt.Println("bcrypt 安全性特点:")
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println("1. 自适应: cost值越高，计算越慢，抵抗暴力破解")
	fmt.Println("2. 内置盐值: 每次生成的哈希都不同，防止彩虹表攻击")
	fmt.Println("3. 单向性: 无法从哈希值反推原密码")
	fmt.Println("4. 工作因子: 随着硬件性能提升，可以增加cost值")
	fmt.Println("\n推荐:")
	fmt.Println("- 生产环境建议 cost=12 或更高")
	fmt.Println("- 定期评估并调整cost值")
	fmt.Println("- 对于更高安全性，考虑使用 Argon2id")
}
