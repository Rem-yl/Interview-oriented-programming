package main

import (
	"fmt"
	"time"
)

var Version = "0.0"
var BuildTime = time.Now().String()

func main() {
	fmt.Printf("Version: %s \n", Version)
	fmt.Printf("BuildTime: %s \n", BuildTime)
	fmt.Println("Welcome to load balancer project!")
}
