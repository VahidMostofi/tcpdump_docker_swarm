package main

import (
	"fmt"
	"os"
)

func executeExtract(overlayNetworkName string) {
	d := ExtractInformation(os.Args[2])
	d.Save()
	fmt.Println("saved under name:", d.Networks["overlay"].ShortID)
}

func executeTCPDUMPCMD(name string) {
	d := LoadDeploymentInfo(name)
	RunTCPDUMP(d)
}

func executeParse(name string) {
	d := LoadDeploymentInfo(name)
	Parse(d)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("use one of these commands:")
		fmt.Println("extract <network_name>")
		fmt.Println("tcpdump <dir_name>")
		fmt.Println("tcpdump <dir_name>")
	}
	if os.Args[1] == "extract" {
		executeExtract(os.Args[2])
	} else if os.Args[1] == "tcpdump" {
		if len(os.Args) != 3 {
			fmt.Println("you must provide a name")
			fmt.Println("tcpdump <dir_name>")
			return
		}
		executeTCPDUMPCMD(os.Args[2])
	} else if os.Args[1] == "parse" {
		if len(os.Args) != 3 {
			fmt.Println("you must provide a name")
			fmt.Println("tcpdump <dir_name>")
			return
		}
		executeParse(os.Args[2])
	}
}
