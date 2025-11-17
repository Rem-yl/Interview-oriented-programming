package test

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	hashfuncs "hash-learn/hash_funcs"
	"testing"
)

func TestDGB2(t *testing.T) {
	hashFunc := hashfuncs.NewDGB2()

	s := "hello, world"
	value, _ := hashFunc.Sum([]byte(s))
	hexStr := hex.EncodeToString(value)
	fmt.Println(hexStr)
}

func TestFNV(t *testing.T) {
	hashFunc := hashfuncs.NewFNV()

	s := "hello, world"
	value, _ := hashFunc.Sum([]byte(s))
	hexStr := hex.EncodeToString(value)
	fmt.Println(hexStr)
}

func TestMD5(t *testing.T) {
	hashFunc := hashfuncs.NewMD5()

	s := "hello, world"
	value, _ := hashFunc.Sum([]byte(s))
	hexStr := hex.EncodeToString(value)

	md5Hash := md5.Sum([]byte(s))
	md5Str := hex.EncodeToString(md5Hash[:])
	if hexStr != md5Str {
		panic("not same")
	}
}
