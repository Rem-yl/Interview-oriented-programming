package ch0101

// 习题1.1.13: 二维矩阵的转置
func Transport2D(matrix [][]int) [][]int {
	rows := len(matrix)
	if rows < 1 {
		return [][]int{}
	}

	cols := len(matrix[0])

	// 共分配 cols + 1次内存
	matrix_t := make([][]int, cols) // 转置后的矩阵
	for i := range cols {
		matrix_t[i] = make([]int, rows)
	}

	for i := range rows {
		for j := range cols {
			matrix_t[j][i] = matrix[i][j]
		}
	}

	return matrix_t
}

// 习题 1.1.15
func Histogram(a []int, m int) []int {
	res := make([]int, m)

	for _, num := range a {
		res[num]++
	}

	return res
}

// 二项分布
// P(X = k) = C(n,k) * p^k * (1-p)^(n-k)

var binoMem = make(map[[2]int]float64, 5)

func Binomial(n, k int, p float64) float64 {
	if k < 0 || k > n {
		return 0.0
	}
	if n == 0 {
		if k == 0 {
			return 1.0
		}
		return 0.0
	}

	key := [2]int{n, k}
	if val, ok := binoMem[key]; ok {
		return val
	}

	res := (1.0-p)*Binomial(n-1, k, p) + p*Binomial(n-1, k-1, p)
	binoMem[key] = res
	return res
}
