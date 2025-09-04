// 本代码演示了Go语言的基本用法
package main

import (
	"fmt"
	"sync"
	"time"
)

// GoVarTemp 函数演示了 Go 语言的基本类型、变量的定义方式和常量
func GoVarTemp() {
	// ====== 基本类型 ======
	var a int = 10
	var b float64 = 3.14
	var c string = "Hello"
	var d bool = true
	var e byte = 'A' // 字节（uint8 的别名）
	var f rune = '中' // Unicode 码点（int32 的别名）

	// ====== 定义方式 ======
	var x int //默认零值
	var y = 20
	z := 30 // 短变量声明（只能在函数内用）

	// ====== 常量 ======
	const pi = 3.14159
	const greeting string = "hello"

	fmt.Println("a =", a)
	fmt.Println("b =", b)
	fmt.Println("c =", c)
	fmt.Println("d =", d)
	fmt.Println("e =", e, "字符 =", string(e))
	fmt.Println("f =", f, "字符 =", string(f))

	fmt.Println("x =", x)
	fmt.Println("y =", y)
	fmt.Println("z =", z)

	fmt.Println("pi =", pi)
	fmt.Println("greeting =", greeting)
}

type Shape interface {
	CalArea() float64
	CalPerimeter() float64
}

type Rectangle struct {
	Width, Height float64
}

func (r *Rectangle) CalArea() float64 {
	return r.Width * r.Height
}

func (r *Rectangle) CalPerimeter() float64 {
	return 2 * (r.Height + r.Width)
}

type Circle struct {
	Radius float64
}

func (c *Circle) CalArea() float64 {
	return 3.14 * c.Radius * c.Radius
}

func (c *Circle) CalPerimeter() float64 {
	return 2 * 3.14 * c.Radius
}

func printArea(s Shape) {
	fmt.Printf("Area is: %.2f\n", s.CalArea())
}

func SturctTemp() {
	r := &Rectangle{
		Width: 0.5, Height: 0.3,
	}
	c := &Circle{
		Radius: 0.8,
	}

	printArea(r)
	printArea(c)
}

func printNumbers(name string, wg *sync.WaitGroup) {
	defer wg.Done()

	for i := 0; i <= 10; i++ {
		fmt.Printf("%s: %d\n", name, i)
		time.Sleep(100 * time.Millisecond)
	}
}

func GoRoutineTemp() {
	var wg sync.WaitGroup

	wg.Add(2)
	fmt.Println("Start Print Number: ")
	go printNumbers("thread 1", &wg)
	go printNumbers("thread 2", &wg)

	wg.Wait()
	fmt.Println("All Done.")
}

func selectTicker() {
	tick1 := time.NewTicker(1 * time.Second)
	tick2 := time.NewTicker(2 * time.Second)

	stop := time.After(5 * time.Second)

	for {
		select {
		case <-tick1.C:
			fmt.Println("Tick1")
		case <-tick2.C:
			fmt.Println("Tick2")
		case <-stop:
			fmt.Println("Stop")
			tick1.Stop()
			tick2.Stop()
			return
		}
	}
}

func main() {
	// GoVarTemp()
	// SturctTemp()
	GoRoutineTemp()
	selectTicker()
}
