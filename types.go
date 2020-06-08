package main

import (
	"encoding/json"
	"io/ioutil"
)

var FSBase = "/home/vahid/Desktop/temp1/tcpdumps"

type DeploymentInfo struct {
	DNS       map[string]string `json:"dns"`
	NetworkID string            `json:"networkID"`
}

func (d *DeploymentInfo) Save() {
	b, err := json.MarshalIndent(d, "", "\t")
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile(FSBase+"/"+d.NetworkID[:12]+".json", b, 0644)
}

func LoadDeploymentInfo(name string) *DeploymentInfo {
	d := DeploymentInfo{}
	b, err := ioutil.ReadFile(FSBase + "/" + name + ".json")
	if err != nil {
		panic(err)
	}
	json.Unmarshal(b, &d)
	return &d
}
