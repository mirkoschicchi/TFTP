package main

import (
	"fmt"
)

func main() {
	fmt.Println("Welcome! Insert the IP address of the machine you want to connect to...")
	var targetIpAddress string
	fmt.Scanln(&targetIpAddress)

	fmt.Println("Insert the port of the target address running the TFTP server...")
	var targetPort int
	fmt.Scanln(&targetPort)
}
