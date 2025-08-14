package main

import (
	"fmt"
	"net"
	"os"
)

func lookHostName(hostname string) ([]string, error) {
	IPs, err := net.LookupHost(hostname)

	if err != nil {
		return nil, err
	}

	return IPs, nil
}

func lookIP(address string) ([]string, error) {
	hosts, err := net.LookupAddr(address)
	if err != nil {
		return nil, err
	}

	return hosts, nil
}

func main() {
	args := os.Args
	if len(args) == 1 {
		fmt.Println("Please provide host")
		return
	}

	input := args[1]
	IPaddress := net.ParseIP(input)

	if IPaddress == nil {
		IPs, err := lookHostName(input)
		if err == nil {
			for _, singleIP := range IPs {
				fmt.Println(singleIP)
			}
		}
	} else {
		hosts, err := lookIP(input)
		if err == nil {
			for _, hostname := range hosts {
				fmt.Println(hostname)
			}
		}
	}
}
