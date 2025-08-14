package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	args := os.Args

	if len(args) == 1 {
		fmt.Println("Please provide name.")
		return
	}

	name := args[1]

	address := "127.0.0.1:8801"

	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Printf("Dial on %s error: %s", address, err)
		return
	}
	defer conn.Close()

	stdScanner := bufio.NewScanner(os.Stdin)
	connScanner := bufio.NewScanner(conn)

	// 用于接受消息的 goroutine
	go func() {
		for connScanner.Scan() {
			text := connScanner.Text()
			fmt.Println(":->", text)
			fmt.Print(">>>")
		}
	}()

	// 主线程用于扫描终端输入
	fmt.Print(">>>")

	for stdScanner.Scan() {
		msg := stdScanner.Text()

		fullMsg := fmt.Sprintf("(%s) - %s\n", name, msg)
		if _, err := fmt.Fprintf(conn, fullMsg); err != nil {
			fmt.Printf("[Main] Write to %s failed \n", conn.RemoteAddr())
			continue
		}
	}
}
