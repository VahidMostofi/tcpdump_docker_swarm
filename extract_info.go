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

// Returns Network from Docker based on networkName and networkDriver. It applies networkDriver's filter if it is anything other than empty string
func getNetwork(cli *client.Client, ctx context.Context, networkName, networkDriver string) (types.NetworkResource, error) {
	networkFilter := filters.NewArgs()
	networkFilter.Add("name", networkName)
	if len(networkDriver) > 0 {
		networkFilter.Add("driver", networkDriver)
	}
	networks, err := cli.NetworkList(ctx, types.NetworkListOptions{
		Filters: networkFilter,
	})

	if err != nil {
		return types.NetworkResource{}, err
	}

	if len(networks) < 1 {
		return types.NetworkResource{}, fmt.Errorf("there is no network with these filters")
	}
	return networks[0], nil
}

// prefix would be added to end of the container name, it also adds the _
func extractDNSInfoFromNetwork(cli *client.Client, ctx context.Context, networkID, prefix string, dns map[string]string) error {
	network, err := cli.NetworkInspect(ctx, networkID)
	if err != nil {
		return err
	}

	for _, endpointResource := range network.Containers {
		var name string
		if strings.Contains(endpointResource.Name, "_") {
			fmt.Println(endpointResource.Name)
			name = strings.Split(endpointResource.Name, "_")[1]
		} else {
			name = endpointResource.Name
		}

		temp := strings.Split(name, ".")
		if len(temp) > 2 {
			name = temp[0] + "." + temp[1]
		} else if len(temp) == 1 {
			name = temp[0]
		}
		name += "_" + prefix
		dns[cleanIp(endpointResource.IPv4Address)] = name
	}
	return nil
}

func ExtractInformation() *DeploymentInfo {
	deploymenInfo := DeploymentInfo{
		DNS: make(map[string]string),
	}

	deploymenInfo.DNS["136.159.209.204"] = "136.159.209.204"
	deploymenInfo.DNS["136.159.209.214"] = "136.159.209.214"
	deploymenInfo.DNS["50.99.77.228"] = "50.99.77.228"
	deploymenInfo.DNS["172.20.0.2"] = "172.20.0.2"

	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	ctx := context.Background()
	cli, err := client.NewClient("tcp://136.159.209.204:2375", "", nil, defaultHeaders)
	if err != nil {
		panic(err)
	}

	defaultNetwork, err := getNetwork(cli, ctx, "bookstore_default", "overlay")
	if err != nil {
		panic(err)
	}
	deploymenInfo.DefaultNetworkID = defaultNetwork.ID

	ingressNetwork, err := getNetwork(cli, ctx, "ingress-network", "")
	if err != nil {
		panic(err)
	}
	deploymenInfo.IngressNetworkID = ingressNetwork.ID

	services, err := cli.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		panic(err)
	}

	// Extracting IP address for each service in each network
	for _, service := range services {
		for _, evip := range service.Endpoint.VirtualIPs {
			if evip.NetworkID == defaultNetwork.ID {
				name := service.Spec.Name
				name = strings.Replace(name, "bookstore_", "", 1)
				name += "_srv_def"
				deploymenInfo.DNS[cleanIp(evip.Addr)] = name
			} else if evip.NetworkID == ingressNetwork.ID {
				name := service.Spec.Name
				name = strings.Replace(name, "bookstore_", "", 1)
				name += "_srv_ing"
				deploymenInfo.DNS[cleanIp(evip.Addr)] = name
			}
		}

	}

	err = extractDNSInfoFromNetwork(cli, ctx, defaultNetwork.ID, "def", deploymenInfo.DNS)
	if err != nil {
		panic(err)
	}
	err = extractDNSInfoFromNetwork(cli, ctx, ingressNetwork.ID, "ing", deploymenInfo.DNS)
	if err != nil {
		panic(err)
	}

	return &deploymenInfo
}
