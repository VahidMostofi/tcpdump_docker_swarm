package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("use one of these commands:")
		fmt.Println("extract <dir_name>")
		fmt.Println("tcpdump <dir_name> <extracted_information_name>")
	}
	if os.Args[1] == "extract" {
		FSBase += "/" + os.Args[2]
		d := ExtractInformation()
		d.Save()
		fmt.Println("saved under name:", d.NetworkID[:12])
	} else if os.Args[1] == "tcpdump" {
		if len(os.Args) < 3 {
			fmt.Println("you must provide name of network")
		}
		name := os.Args[2]
		if len(name) > 12 {
			name = name[:12]
		}
		d := LoadDeploymentInfo(name)
		RunTCPDUMP(d)
	} else if os.Args[1] == "parse" {
		if len(os.Args) < 3 {
			fmt.Println("you must provide name of network")
		}
		name := os.Args[2]
		if len(name) > 12 {
			name = name[:12]
		}
		d := LoadDeploymentInfo(name)
		Parse(d)
	}

	// ExtractInformation()
	// // Parse()
	// // RunTCPDUMP()
}
