package hashfuncs

import (
	"encoding/binary"
	"math"
)

type MD5 struct {
	a, b, c, d uint32
}

// 非线性函数
func fFunc(x, y, z uint32) uint32 { return (x & y) | (^x & z) }
func gFunc(x, y, z uint32) uint32 { return (x & z) | (y & ^z) }
func hFunc(x, y, z uint32) uint32 { return x ^ y ^ z }
func iFunc(x, y, z uint32) uint32 { return y ^ (x | ^z) }

func leftRotate(x, n uint32) uint32 {
	return (x << n) | (x >> (32 - n))
}

func NewMD5() *MD5 {
	return &MD5{
		a: 0x67452301,
		b: 0xEFCDAB89,
		c: 0x98BADCFE,
		d: 0x10325476,
	}
}

// MD5填充: [原始数据] + 0x80 + 0x00...00 + [8 字节长度]
func (h *MD5) pad(data []byte) []byte {
	msgLen := uint64(len(data)) * 8 // 计算 8 字节长度
	data = append(data, 0x80)

	for len(data)%64 != 56 {
		data = append(data, 0x00)
	}

	lenBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(lenBytes, msgLen)
	data = append(data, lenBytes...)

	return data
}

func (h *MD5) processBlock(block []byte) {
	// 1. 分解为16个32位字 (16 * 32 = 512位)
	x := make([]uint32, 16)
	for i := 0; i < 16; i++ {
		x[i] = binary.LittleEndian.Uint32(block[i*4 : (i+1)*4])
	}

	// 2. 预计算T常数表: T[i] = floor(2^32 × |sin(i+1)|)
	T := make([]uint32, 64)
	for i := 0; i < 64; i++ {
		T[i] = uint32(math.Floor(math.Abs(math.Sin(float64(i+1))) * math.Pow(2, 32)))
	}

	// 3. 每轮的循环左移位数
	s := []uint32{
		7, 12, 17, 22, 7, 12, 17, 22, 7, 12, 17, 22, 7, 12, 17, 22, // 第1轮
		5, 9, 14, 20, 5, 9, 14, 20, 5, 9, 14, 20, 5, 9, 14, 20,     // 第2轮
		4, 11, 16, 23, 4, 11, 16, 23, 4, 11, 16, 23, 4, 11, 16, 23, // 第3轮
		6, 10, 15, 21, 6, 10, 15, 21, 6, 10, 15, 21, 6, 10, 15, 21, // 第4轮
	}

	// 4. 初始化工作变量
	a, b, c, d := h.a, h.b, h.c, h.d

	// 5. 主循环：64步运算
	for i := uint32(0); i < 64; i++ {
		var f uint32  // 非线性函数的结果
		var k int     // 消息字的索引

		// 根据轮次选择不同的非线性函数和消息字访问顺序
		if i < 16 {
			// 第1轮: F函数，顺序访问
			f = fFunc(b, c, d)
			k = int(i)
		} else if i < 32 {
			// 第2轮: G函数，跳跃访问
			f = gFunc(b, c, d)
			k = int((5*i + 1) % 16)
		} else if i < 48 {
			// 第3轮: H函数，跳跃访问
			f = hFunc(b, c, d)
			k = int((3*i + 5) % 16)
		} else {
			// 第4轮: I函数，跳跃访问
			f = iFunc(b, c, d)
			k = int((7 * i) % 16)
		}

		// MD5 单步运算:
		// 1. 计算: temp = a + F(b,c,d) + M[k] + T[i]
		// 2. 循环左移: temp = ROTATE_LEFT(temp, s)
		// 3. 加上b: temp = temp + b
		temp := a + f + x[k] + T[i]
		temp = leftRotate(temp, s[i])
		temp = temp + b

		// 状态轮转: (a, b, c, d) ← (d, temp, b, c)
		// 解释:
		//   新的a = 旧的d
		//   新的b = temp
		//   新的c = 旧的b
		//   新的d = 旧的c
		a, b, c, d = d, temp, b, c
	}

	// 6. 累加到原始状态 (防止被覆盖)
	h.a += a
	h.b += b
	h.c += c
	h.d += d
}

func (h *MD5) Sum(data []byte) ([]byte, error) {
	paddedData := h.pad(data)

	// 处理 512 位: 64 * 8 = 512
	for i := 0; i < len(paddedData); i += 64 {
		h.processBlock(paddedData[i : i+64])
	}

	var res [16]byte
	binary.LittleEndian.PutUint32(res[0:4], h.a)
	binary.LittleEndian.PutUint32(res[4:8], h.b)
	binary.LittleEndian.PutUint32(res[8:12], h.c)
	binary.LittleEndian.PutUint32(res[12:16], h.d)

	return res[:], nil
}
