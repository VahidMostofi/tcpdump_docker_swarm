package main

import (
	"archive/tar"
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

var defaultHeaders = map[string]string{"User-Agent": "engine-api-cli-1.0"}

func executeTCPDUMP(cli *client.Client, ctx context.Context, name, key string, network *TCPDUMPNetworkInfo, errc chan error, done chan string) {

	containerName := "net_dbg_" + key
	err := removeContainerByName(cli, ctx, containerName)
	if err != nil {
		errc <- err
	}
	containerCreateBody, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        "nicolaka/netshoot",
		OpenStdin:    true,
		StdinOnce:    true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{"nsenter", "--net=/var/run/docker/netns/" + network.FSName, "sh"},
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "/var/run/docker/netns",
				Target: "/var/run/docker/netns",
			},
		},
		Privileged: true,
	}, nil, containerName)
	if err != nil {
		errc <- err
		return
	} else {
		defer cli.ContainerRemove(ctx, containerCreateBody.ID, types.ContainerRemoveOptions{Force: true})
	}
	err = cli.ContainerStart(ctx, containerCreateBody.ID, types.ContainerStartOptions{})
	if err != nil {
		errc <- err
		return
	} else {
		defer cli.ContainerStop(ctx, containerCreateBody.ID, nil)
	}
	fmt.Println("container started for " + key + ", starting tcpdump...")
	IFConfigCommand := []string{"timeout", "60", "tcpdump", "-i", "any", "-vv", "-X", "-w", network.ShortID}
	resp, err := Exec(ctx, containerCreateBody.ID, IFConfigCommand)
	if err != nil {
		errc <- err
		return
	}
	execRes, err := InspectExecResp(ctx, resp.ID)
	if err != nil {
		errc <- err
		return
	}
	if len(execRes.StdOut) > 0 {
		fmt.Println(key + ": " + execRes.StdOut)
	}
	if len(execRes.StdErr) > 0 {
		fmt.Println(key + ": " + execRes.StdErr)
	}

	readCloser, _, err := cli.CopyFromContainer(ctx, containerCreateBody.ID, network.ShortID)
	if err != nil {
		errc <- err
		return
	}
	defer readCloser.Close()

	tarHandle := tar.NewReader(readCloser)
	_, err = tarHandle.Next()
	if err != nil {
		errc <- err
		return
	}
	b, err := ioutil.ReadAll(tarHandle)
	if err != nil {
		errc <- err
		return
	}
	err = ioutil.WriteFile(FSBase+"/"+name+"/"+network.ShortID+".pcap", b, 0777)
	if err != nil {
		errc <- err
		return
	}
	fmt.Println("finished copying")
	done <- FSBase + "/" + name + "/" + network.ShortID + ".pcap"
}

func RunTCPDUMP(deploymentInfo *DeploymentInfo) {
	networks := deploymentInfo.Networks
	name := networks["overlay"].ShortID
	ctx := context.Background()
	cli, err := client.NewClient("tcp://136.159.209.204:2375", "", nil, defaultHeaders)
	if err != nil {
		panic(err)
	}
	errc := make(chan error)
	done := make(chan string)
	fmt.Println(networks)
	for key, network := range networks {
		fmt.Println("calling with ", key)
		go executeTCPDUMP(cli, ctx, name, key, network, errc, done)
	}
	count := 0
	argsToMerge := []string{"-a"}
	for {
		select {
		case err := <-errc:
			panic(err)
		case fileName := <-done:
			count++
			argsToMerge = append(argsToMerge, fileName)
			if count == len(networks) {
				time.Sleep(time.Second * 4)
				argsToMerge = append(argsToMerge, []string{"-w", FSBase + "/" + name + "/merged.pcap"}...)
				cmd := exec.Command("mergecap", argsToMerge...)
				err := cmd.Run()
				if err != nil {
					panic(err)
				}
				return
			}
		}
	}

}
