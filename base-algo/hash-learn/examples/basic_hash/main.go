package main

import (
	"crypto/sha256"
	"fmt"
)

// DJB2 哈希算法
func djb2(s string) uint32 {
	hash := uint32(5381)
	for _, c := range s {
		hash = ((hash << 5) + hash) + uint32(c) // hash * 33 + c
	}
	return hash
}

const (
	fnvOffsetBasis32 = 2166136261
	fnvPrime32       = 16777619
	fnvOffsetBasis64 = 14695981039346656037
	fnvPrime64       = 1099511628211
)

// FNV-1a 32位哈希
func fnv1a32(data []byte) uint32 {
	hash := uint32(fnvOffsetBasis32)
	for _, b := range data {
		hash ^= uint32(b)
		hash *= fnvPrime32
	}
	return hash
}

// FNV-1a 64位哈希
func fnv1a64(data []byte) uint64 {
	hash := uint64(fnvOffsetBasis64)
	for _, b := range data {
		hash ^= uint64(b)
		hash *= fnvPrime64
	}
	return hash
}

func main() {
	testStrings := []string{"hello", "world", "golang", "hash"}

	// DJB2 演示
	fmt.Println("=" + string(make([]byte, 60)) + "=")
	fmt.Println("DJB2 Hash Demo:")
	fmt.Println("=" + string(make([]byte, 60)) + "=")
	for _, s := range testStrings {
		fmt.Printf("djb2(\"%s\") = %d (0x%08x)\n", s, djb2(s), djb2(s))
	}

	// 演示雪崩效应
	fmt.Println("\n雪崩效应演示:")
	fmt.Printf("djb2(\"hello\") = 0x%08x\n", djb2("hello"))
	fmt.Printf("djb2(\"hallo\") = 0x%08x\n", djb2("hallo"))

	// FNV-1a 演示
	fmt.Println("\n" + "=" + string(make([]byte, 60)) + "=")
	fmt.Println("FNV-1a Hash Demo:")
	fmt.Println("=" + string(make([]byte, 60)) + "=")
	for _, s := range testStrings {
		data := []byte(s)
		fmt.Printf("fnv1a32(\"%s\") = 0x%08x\n", s, fnv1a32(data))
		fmt.Printf("fnv1a64(\"%s\") = 0x%016x\n", s, fnv1a64(data))
		fmt.Println()
	}

	// SHA-256 演示
	fmt.Println("=" + string(make([]byte, 60)) + "=")
	fmt.Println("SHA-256 Hash Demo:")
	fmt.Println("=" + string(make([]byte, 60)) + "=")

	testData := []string{
		"hello",
		"Hello", // 大小写变化
		"hello world",
		"The quick brown fox jumps over the lazy dog",
	}

	for _, s := range testData {
		hash := sha256.Sum256([]byte(s))
		fmt.Printf("SHA256(\"%s\")\n  = %x\n\n", s, hash)
	}

	// SHA-256 雪崩效应演示
	fmt.Println("SHA-256 雪崩效应演示:")
	h1 := sha256.Sum256([]byte("hello"))
	h2 := sha256.Sum256([]byte("hallo"))
	fmt.Printf("SHA256(\"hello\") = %x\n", h1)
	fmt.Printf("SHA256(\"hallo\") = %x\n", h2)

	// 计算不同位的数量
	diff := 0
	for i := 0; i < 32; i++ {
		xor := h1[i] ^ h2[i]
		for j := 0; j < 8; j++ {
			if (xor & (1 << j)) != 0 {
				diff++
			}
		}
	}
	fmt.Printf("改变1个字符，导致%d位改变（共256位，%.1f%%）\n", diff, float64(diff)/256*100)
}
