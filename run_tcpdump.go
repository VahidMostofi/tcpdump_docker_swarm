package main

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

var defaultHeaders = map[string]string{"User-Agent": "engine-api-cli-1.0"}

func removeContainerByName(cli *client.Client, ctx context.Context, containerName string) error {
	containerListFilters := filters.NewArgs()
	containerListFilters.Add("name", containerName)
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: containerListFilters})
	if err != nil {
		return err
	}
	if len(containers) == 0 {
		return nil
	}
	for _, c := range containers {
		cli.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{Force: true})
	}
	return nil
}

func executeTCPDUMP(cli *client.Client, ctx context.Context, key, networkID string, errc chan error, done chan string) {
	var networkFileName string
	if networkID == "default" {
		networkFileName = "default"
	} else {
		networkFileName = "1-" + networkID[:10]
	}

	prettyNetworkID := networkID
	if len(prettyNetworkID) > 12 {
		prettyNetworkID = prettyNetworkID[:12]
	}
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
		Cmd:          []string{"nsenter", "--net=/var/run/docker/netns/" + networkFileName, "sh"},
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
	IFConfigCommand := []string{"timeout", "60", "tcpdump", "-i", "any", "-vv", "-X", "-w", prettyNetworkID}
	resp, err := Exec(ctx, containerCreateBody.ID, IFConfigCommand)
	if err != nil {
		panic(err)
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

	readCloser, _, err := cli.CopyFromContainer(ctx, containerCreateBody.ID, prettyNetworkID)
	if err != nil {
		panic(err)
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
	err = ioutil.WriteFile(FSBase+"/"+prettyNetworkID+".pcap", b, 0777)
	if err != nil {
		errc <- err
		return
	}
	fmt.Println("finished copying")
	done <- FSBase + "/" + prettyNetworkID + ".pcap"
}

func RunTCPDUMP(deploymentInfo *DeploymentInfo) {
	networks := map[string]string{"default": deploymentInfo.DefaultNetworkID, "ingress": deploymentInfo.IngressNetworkID, "host": "default"}

	ctx := context.Background()
	cli, err := client.NewClient("tcp://136.159.209.204:2375", "", nil, defaultHeaders)
	if err != nil {
		panic(err)
	}
	errc := make(chan error)
	done := make(chan string)
	fmt.Println(networks)
	for key, networkID := range networks {
		fmt.Println("calling with ", networkID)
		go executeTCPDUMP(cli, ctx, key, networkID, errc, done)
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
				argsToMerge = append(argsToMerge, []string{"-w", FSBase + "/merged.pcap"}...)
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

type ExecResult struct {
	StdOut   string
	StdErr   string
	ExitCode int
}

func Exec(ctx context.Context, containerID string, command []string) (types.IDResponse, error) {
	docker, err := client.NewClient("tcp://136.159.209.204:2375", "", nil, defaultHeaders)
	if err != nil {
		return types.IDResponse{}, err
	}
	docker.Close()

	config := types.ExecConfig{
		AttachStderr: true,
		AttachStdout: true,
		Tty:          true,
		Cmd:          command,
	}

	return docker.ContainerExecCreate(ctx, containerID, config)
}

func InspectExecResp(ctx context.Context, id string) (ExecResult, error) {
	var execResult ExecResult
	docker, err := client.NewClient("tcp://136.159.209.204:2375", "", nil, defaultHeaders)
	if err != nil {
		return execResult, err
	}
	defer docker.Close()

	resp, err := docker.ContainerExecAttach(ctx, id, types.ExecConfig{})
	if err != nil {
		return execResult, err
	}

	defer resp.Close()

	// read the output
	var outBuf, errBuf bytes.Buffer
	outputDone := make(chan error)

	go func() {
		// StdCopy demultiplexes the stream into two buffers
		_, err = stdcopy.StdCopy(&outBuf, &errBuf, resp.Reader)
		outputDone <- err
	}()

	select {
	case err := <-outputDone:
		if err != nil {
			return execResult, err
		}
		break

	case <-ctx.Done():
		return execResult, ctx.Err()
	}

	stdout, err := ioutil.ReadAll(&outBuf)
	if err != nil {
		return execResult, err
	}
	stderr, err := ioutil.ReadAll(&errBuf)
	if err != nil {
		return execResult, err
	}

	res, err := docker.ContainerExecInspect(ctx, id)
	if err != nil {
		return execResult, err
	}

	execResult.ExitCode = res.ExitCode
	execResult.StdOut = string(stdout)
	execResult.StdErr = string(stderr)
	return execResult, nil
}
