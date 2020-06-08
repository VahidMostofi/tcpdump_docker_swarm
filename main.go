package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("use one of these commands:")
		fmt.Println("extract")
		fmt.Println("tcpdump <dir_name> <extracted_information_name>")
		fmt.Println("tcpdump <dir_name>")
	}
	if os.Args[1] == "extract" {
		d := ExtractInformation()
		d.Save()
		fmt.Println("saved under name:", d.DefaultNetworkID[:12])
	} else if os.Args[1] == "tcpdump" {
		if len(os.Args) < 3 {
			fmt.Println("you must provide a directory name and name of network")
			fmt.Println("tcpdump <dir_name>")
			return
		}
		FSBase += "/" + os.Args[2]
		d := LoadDeploymentInfo()
		RunTCPDUMP(d)
	} else if os.Args[1] == "parse" {
		if len(os.Args) < 3 {
			fmt.Println("you must provide name of network")
		}
		FSBase += "/" + os.Args[2]
		d := LoadDeploymentInfo()
		Parse(d)
	}

	// ExtractInformation()
	// // Parse()
	// // RunTCPDUMP()
}
