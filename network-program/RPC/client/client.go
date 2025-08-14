// go run ../server/server.gp
// go run client.go

package main

import (
	"RPC/shared"
	"fmt"
	"net/rpc"
)

func main() {
	address := "127.0.0.1:8848"

	conn, err := rpc.Dial("tcp", address)
	if err != nil {
		fmt.Printf("Dial %s failed.", address)
		return
	}

	var num1, num2 float64
	for {
		fmt.Println("请输出两个数字(使用空格分隔): ")
		if _, err := fmt.Scan(&num1, &num2); err != nil {
			fmt.Println("输入错误，请重新输入:", err)
			continue
		}

		arg := shared.Data{
			Num1: num1, Num2: num2,
		}

		var res float64
		if err := conn.Call("RPCServer.Mul", arg, &res); err != nil {
			fmt.Println(err)
		}

		fmt.Printf("Mul(%f, %f) = %f \n", num1, num2, res)
	}

}
