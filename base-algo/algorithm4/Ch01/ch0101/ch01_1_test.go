package ch0101

import (
	"reflect"
	"testing"
)

func TestTransport2D(t *testing.T) {
	tests := []struct {
		name  string
		input [][]int
		want  [][]int
	}{
		{"空数组", [][]int{}, [][]int{}},
		{"单行数组", [][]int{{1, 2, 3}}, [][]int{{1}, {2}, {3}}},
		{"单列数组", [][]int{{1}, {2}, {3}}, [][]int{{1, 2, 3}}},
		{"多列数据", [][]int{{1, 2, 3}, {4, 5, 6}}, [][]int{{1, 4}, {2, 5}, {3, 6}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Transport2D(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// 基准测试 - 表驱动模式
func BenchmarkTransport2D(b *testing.B) {
	benchmarks := []struct {
		name string
		rows int
		cols int
	}{
		{"小矩阵-10x10", 10, 10},
		{"中矩阵-100x100", 100, 100},
		{"大矩阵-1000x1000", 1000, 1000},
		{"长方形-100x1000", 100, 1000},
		{"长方形-1000x100", 1000, 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			// 准备测试数据
			matrix := makeMatrix(bm.rows, bm.cols)

			b.ReportAllocs() // 报告内存分配
			b.ResetTimer()   // 重置计时器，排除初始化时间

			var result [][]int
			for i := 0; i < b.N; i++ {
				result = Transport2D(matrix)
			}
			_ = result // 防止编译器优化掉结果
		})
	}
}

// 辅助函数：生成测试矩阵
func makeMatrix(rows, cols int) [][]int {
	matrix := make([][]int, rows)
	for i := range matrix {
		matrix[i] = make([]int, cols)
		for j := range matrix[i] {
			matrix[i][j] = i*cols + j
		}
	}
	return matrix
}

func TestHistogram(t *testing.T) {
	tests := []struct {
		name  string
		m     int
		input []int
		want  []int
	}{
		{"m5n10", 5, []int{2, 1, 2, 0, 0, 3, 2, 1, 4, 4}, []int{2, 2, 3, 1, 2}},
		{"m4n4", 4, []int{1, 2, 1, 0}, []int{1, 2, 1, 0}},
		{"全零", 4, []int{0, 0, 0, 0}, []int{4, 0, 0, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Histogram(tt.input, tt.m)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got: %v, want: %v", got, tt.want)
			}

		})
	}
}
