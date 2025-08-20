// 展示flag基本用法

package main

import (
	"flag"
	"log"
)

func test1() {
	var name string
	var age int

	flag.StringVar(&name, "name", "", "name")
	flag.IntVar(&age, "age", 0, "age")

	flag.Parse()

	log.Printf("name: %s, age: %d \n", name, age)
}

func test2() {
	flag.Parse()

	args := flag.Args()

	if len(args) <= 0 {
		return
	}

	var name string

	switch args[0] {
	case "rem":
		remCmd := flag.NewFlagSet("rem", flag.ExitOnError)
		remCmd.StringVar(&name, "name", "", "rem name")
		_ = remCmd.Parse(args[1:])
	case "ram":
		ramCmd := flag.NewFlagSet("ram", flag.ExitOnError)
		ramCmd.StringVar(&name, "name", "", "ram name")
		_ = ramCmd.Parse(args[1:])
	}

	log.Printf("name: %s \n", name)
}

func main() {
	test2()
}
