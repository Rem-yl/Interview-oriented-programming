package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math"
)

// MD5 简化实现（教学用途）
type MD5 struct {
	a, b, c, d uint32
}

// 非线性函数
func fFunc(x, y, z uint32) uint32 { return (x & y) | (^x & z) }
func gFunc(x, y, z uint32) uint32 { return (x & z) | (y & ^z) }
func hFunc(x, y, z uint32) uint32 { return x ^ y ^ z }
func iFunc(x, y, z uint32) uint32 { return y ^ (x | ^z) }

// 循环左移
func leftRotate(x, n uint32) uint32 {
	return (x << n) | (x >> (32 - n))
}

// 初始化
func NewMD5() *MD5 {
	return &MD5{
		a: 0x67452301,
		b: 0xEFCDAB89,
		c: 0x98BADCFE,
		d: 0x10325476,
	}
}

// 填充消息
func (m *MD5) pad(data []byte) []byte {
	msgLen := uint64(len(data)) * 8
	data = append(data, 0x80) // 附加1位和7个0位

	// 填充0直到长度≡448 (mod 512)
	for len(data)%64 != 56 {
		data = append(data, 0x00)
	}

	// 附加原始长度（64位小端序）
	lenBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(lenBytes, msgLen)
	data = append(data, lenBytes...)

	return data
}

// 处理单个512位块
func (m *MD5) processBlock(block []byte) {
	// 分解为16个32位字（小端序）
	x := make([]uint32, 16)
	for i := 0; i < 16; i++ {
		x[i] = binary.LittleEndian.Uint32(block[i*4 : (i+1)*4])
	}

	// 保存当前状态
	aa, bb, cc, dd := m.a, m.b, m.c, m.d

	// 预计算的正弦表（64个值）
	k := make([]uint32, 64)
	for i := 0; i < 64; i++ {
		k[i] = uint32(math.Floor(math.Abs(math.Sin(float64(i+1))) * math.Pow(2, 32)))
	}

	// 每轮的左移位数
	s := []uint32{
		7, 12, 17, 22, 7, 12, 17, 22, 7, 12, 17, 22, 7, 12, 17, 22,
		5, 9, 14, 20, 5, 9, 14, 20, 5, 9, 14, 20, 5, 9, 14, 20,
		4, 11, 16, 23, 4, 11, 16, 23, 4, 11, 16, 23, 4, 11, 16, 23,
		6, 10, 15, 21, 6, 10, 15, 21, 6, 10, 15, 21, 6, 10, 15, 21,
	}

	// 主循环：64步
	for i := uint32(0); i < 64; i++ {
		var fVal, gVal uint32

		if i < 16 {
			fVal = fFunc(bb, cc, dd)
			gVal = i
		} else if i < 32 {
			fVal = gFunc(bb, cc, dd)
			gVal = (5*i + 1) % 16
		} else if i < 48 {
			fVal = hFunc(bb, cc, dd)
			gVal = (3*i + 5) % 16
		} else {
			fVal = iFunc(bb, cc, dd)
			gVal = (7 * i) % 16
		}

		temp := dd
		dd = cc
		cc = bb
		bb = bb + leftRotate(aa+fVal+k[i]+x[gVal], s[i])
		aa = temp
	}

	// 加上原始值
	m.a += aa
	m.b += bb
	m.c += cc
	m.d += dd
}

// 计算MD5
func (m *MD5) Sum(data []byte) [16]byte {
	// 填充
	paddedData := m.pad(data)

	// 处理每个512位块
	for i := 0; i < len(paddedData); i += 64 {
		m.processBlock(paddedData[i : i+64])
	}

	// 输出结果（小端序）
	var result [16]byte
	binary.LittleEndian.PutUint32(result[0:4], m.a)
	binary.LittleEndian.PutUint32(result[4:8], m.b)
	binary.LittleEndian.PutUint32(result[8:12], m.c)
	binary.LittleEndian.PutUint32(result[12:16], m.d)

	return result
}

func main() {
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println("简化版MD5实现演示")
	fmt.Println("=" + string(make([]byte, 70)) + "=")

	testData := []string{
		"",
		"a",
		"abc",
		"message digest",
		"abcdefghijklmnopqrstuvwxyz",
		"The quick brown fox jumps over the lazy dog",
	}

	fmt.Println("\n自实现MD5哈希结果:\n")

	for _, s := range testData {
		customMD5 := NewMD5()
		hash := customMD5.Sum([]byte(s))
		fmt.Printf("MD5(\"%s\")\n  = %x\n", s, hash)
	}

	// 与标准库对比
	fmt.Println("\n" + "=" + string(make([]byte, 70)) + "=")
	fmt.Println("与Go标准库crypto/md5对比")
	fmt.Println("=" + string(make([]byte, 70)) + "=")

	for _, s := range testData {
		customMD5 := NewMD5()
		customHash := customMD5.Sum([]byte(s))

		stdHash := md5.Sum([]byte(s))

		match := "✓"
		if customHash != stdHash {
			match = "✗"
		}

		fmt.Printf("\n输入: \"%s\"\n", s)
		fmt.Printf("自实现: %x\n", customHash)
		fmt.Printf("标准库: %x\n", stdHash)
		fmt.Printf("匹配: %s\n", match)
	}

	// 演示MD5的特性
	fmt.Println("\n" + "=" + string(make([]byte, 70)) + "=")
	fmt.Println("MD5算法特性演示")
	fmt.Println("=" + string(make([]byte, 70)) + "=")

	// 1. 确定性
	fmt.Println("\n1. 确定性 - 相同输入产生相同输出:")
	input := "hello world"
	for i := 0; i < 3; i++ {
		m := NewMD5()
		hash := m.Sum([]byte(input))
		fmt.Printf("  第%d次: %x\n", i+1, hash)
	}

	// 2. 雪崩效应
	fmt.Println("\n2. 雪崩效应 - 微小改变导致完全不同的输出:")
	h1 := md5.Sum([]byte("hello"))
	h2 := md5.Sum([]byte("hallo"))
	fmt.Printf("  MD5(\"hello\") = %x\n", h1)
	fmt.Printf("  MD5(\"hallo\") = %x\n", h2)

	diff := 0
	for i := 0; i < 16; i++ {
		xor := h1[i] ^ h2[i]
		for j := 0; j < 8; j++ {
			if (xor & (1 << j)) != 0 {
				diff++
			}
		}
	}
	fmt.Printf("  改变1个字符，导致%d位改变（共128位，%.1f%%）\n", diff, float64(diff)/128*100)

	// 3. 固定输出长度
	fmt.Println("\n3. 固定输出长度 - 无论输入多长，输出都是128位:")
	inputs := []string{
		"a",
		"hello world",
		"The quick brown fox jumps over the lazy dog. " +
			"The quick brown fox jumps over the lazy dog. " +
			"The quick brown fox jumps over the lazy dog.",
	}
	for _, input := range inputs {
		hash := md5.Sum([]byte(input))
		fmt.Printf("  输入长度: %3d, 输出: %x (%d位)\n", len(input), hash, len(hash)*8)
	}

	// MD5的安全性说明
	fmt.Println("\n" + "=" + string(make([]byte, 70)) + "=")
	fmt.Println("MD5安全性说明")
	fmt.Println("=" + string(make([]byte, 70)) + "=")
	fmt.Println("⚠️  MD5已被证明不安全，不应用于安全相关场景")
	fmt.Println("\n已知攻击:")
	fmt.Println("- 2004年: 王小云团队发现MD5碰撞攻击方法")
	fmt.Println("- 2008年: 使用MD5伪造SSL证书")
	fmt.Println("- 2012年: Flame恶意软件利用MD5碰撞")
	fmt.Println("\n仍可用于:")
	fmt.Println("✓ 非安全的完整性校验（文件去重等）")
	fmt.Println("✓ 非加密的哈希表")
	fmt.Println("✓ 快速校验和")
	fmt.Println("\n不应用于:")
	fmt.Println("✗ 密码存储")
	fmt.Println("✗ 数字签名")
	fmt.Println("✗ 安全证书")
	fmt.Println("✗ 任何安全相关场景")
	fmt.Println("\n推荐替代方案:")
	fmt.Println("- 数据完整性: SHA-256, SHA-512, BLAKE2")
	fmt.Println("- 密码存储: Argon2id, bcrypt, scrypt")
	fmt.Println("- 数字签名: SHA-256/SHA-512 + RSA/ECDSA")
}
