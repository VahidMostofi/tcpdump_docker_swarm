package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

func cleanIp(ip string) string {
	//TODO what the fuck?! this need be chagned why only these 4 values?
	res := ip
	res = strings.Replace(res, "/8", "", 1)
	res = strings.Replace(res, "/16", "", 1)
	res = strings.Replace(res, "/24", "", 1)
	res = strings.Replace(res, "/32", "", 1)

	return res
}

func ExtractInformation() *DeploymentInfo {
	deploymenInfo := DeploymentInfo{
		DNS: make(map[string]string),
	}

	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	ctx := context.Background()
	cli, err := client.NewClient("tcp://136.159.209.204:2375", "", nil, defaultHeaders)
	if err != nil {
		panic(err)
	}

	networkFilter := filters.NewArgs()
	networkFilter.Add("name", "bookstore_default")
	networkFilter.Add("driver", "overlay")
	networks, err := cli.NetworkList(ctx, types.NetworkListOptions{
		Filters: networkFilter,
	})

	if len(networks) < 1 {
		panic("there is no network with these filters")
	}
	overlayNetwork := networks[0]
	// fmt.Println(overlayNetwork)

	services, err := cli.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		panic(err)
	}

	for _, service := range services {
		for _, evip := range service.Endpoint.VirtualIPs {
			if evip.NetworkID == overlayNetwork.ID {
				name := service.Spec.Name
				name = strings.Replace(name, "bookstore_", "", 1)
				name += "_service"
				fmt.Println(name, cleanIp(evip.Addr)) // this would the ip for each service
				deploymenInfo.DNS[cleanIp(evip.Addr)] = name
			}
		}

	}
	overlayNetwork, err = cli.NetworkInspect(ctx, overlayNetwork.ID)
	if err != nil {
		panic(err)
	}
	deploymenInfo.NetworkID = overlayNetwork.ID
	for _, endpointResource := range overlayNetwork.Containers { // the key here is container id TODO: keep it somewhere
		name := strings.Split(endpointResource.Name, "_")[1]
		temp := strings.Split(name, ".")
		if len(temp) > 2 {
			name = temp[0] + "." + temp[1]
		} else if len(temp) == 1 {
			name = temp[0]
		}

		fmt.Println(name, cleanIp(endpointResource.IPv4Address))
		deploymenInfo.DNS[cleanIp(endpointResource.IPv4Address)] = name
	}

	return &deploymenInfo
}
