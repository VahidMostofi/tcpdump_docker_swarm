package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var FSBase = "/home/vahid/Desktop/temp1/tcpdumps"

type DeploymentInfo struct {
	DNS              map[string]string `json:"dns"`
	DefaultNetworkID string            `json:"default_networkID"`
	IngressNetworkID string            `json:"ingress_networkID"`
}

func (d *DeploymentInfo) Save() {
	b, err := json.MarshalIndent(d, "", "\t")
	if err != nil {
		panic(err)
	}
	direcotry_name := d.DefaultNetworkID[:12]
	if _, err := os.Stat(FSBase + "/" + direcotry_name); os.IsNotExist(err) {
		os.Mkdir(FSBase+"/"+direcotry_name, os.ModeDir|0777)
	}
	ioutil.WriteFile(FSBase+"/"+direcotry_name+"/info.json", b, 0666)
}

func LoadDeploymentInfo() *DeploymentInfo {
	d := DeploymentInfo{}
	b, err := ioutil.ReadFile(FSBase + "/info.json")
	if err != nil {
		panic(err)
	}
	json.Unmarshal(b, &d)
	return &d
}
