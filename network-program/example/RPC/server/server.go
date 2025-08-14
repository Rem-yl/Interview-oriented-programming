package main

import (
	"fmt"
	"math"
	"net"
	"net/rpc"

	"RPC/shared"
)

type RPCServer struct{}

func (r *RPCServer) Mul(args shared.Data, res *float64) error {
	fmt.Printf("Mul(%f, %f) \n", args.Num1, args.Num2)
	*res = args.Num1 * args.Num2
	return nil
}

func (r *RPCServer) Power(args shared.Data, res *float64) error {
	fmt.Printf("Power(%f, %f) \n", args.Num1, args.Num2)
	*res = math.Pow(args.Num1, args.Num2)

	return nil
}

func main() {
	address := "127.0.0.1:8848"

	server := new(RPCServer)
	err := rpc.Register(server)
	if err != nil {
		fmt.Println("RPC server register failed.")
		return
	}

	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("Listen on %s failed.", address)
		return
	}

	fmt.Printf("Listen on %s \n", address)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error: ", err)
		}

		go rpc.ServeConn(conn)
	}

}
