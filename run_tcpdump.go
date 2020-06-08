package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

var defaultHeaders = map[string]string{"User-Agent": "engine-api-cli-1.0"}

func RunTCPDUMP(deploymentInfo *DeploymentInfo) {
	networkName := deploymentInfo.NetworkID
	ctx := context.Background()
	cli, err := client.NewClient("tcp://136.159.209.204:2375", "", nil, defaultHeaders)
	if err != nil {
		panic(err)
	}

	CleanUP(cli, ctx)

	// networkVolumes := map[string]struct{}{"/var/run/docker/netns": "/var/run/docker/netns"}
	containerCreateBody, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        "nicolaka/netshoot",
		OpenStdin:    true,
		StdinOnce:    true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{"nsenter", "--net=/var/run/docker/netns/1-" + networkName[:10], "sh"},
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "/var/run/docker/netns",
				Target: "/var/run/docker/netns",
			},
		},
		Privileged: true,
		//autoremove? TODO
	}, nil, "network_debug")

	if err != nil {
		panic(err)
	}

	fmt.Println("network_debug created", containerCreateBody.ID[:12])

	err = cli.ContainerStart(ctx, containerCreateBody.ID, types.ContainerStartOptions{})
	if err != nil {
		panic(err)
	}

	fmt.Println("network_debug started")
	fmt.Println("starting tcpdump...")
	IFConfigCommand := []string{"timeout", "60", "tcpdump", "-i", "any", "-vv", "-X", "-w", networkName[:12]}
	resp, err := Exec(ctx, containerCreateBody.ID, IFConfigCommand)
	if err != nil {
		panic(err)
	}
	execRes, err := InspectExecResp(ctx, resp.ID)
	fmt.Println(execRes)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(execRes)
	}

	readCloser, containerPathState, err := cli.CopyFromContainer(ctx, containerCreateBody.ID, networkName[:12])
	if err != nil {
		panic(err)
	}
	defer readCloser.Close()
	fmt.Println(containerPathState)

	f, err := os.Create(FSBase + "/" + networkName[:12] + ".tar.xz")
	n, err := io.Copy(f, readCloser)
	fmt.Println(n)
	err2 := StopAndRemoveContainer(cli, ctx, containerCreateBody.ID)
	if err2 != nil {
		panic(err2)
	}
}

func CleanUP(cli *client.Client, ctx context.Context) {
	filters := filters.NewArgs()
	filters.Add("name", "network_debug")
	list, err := cli.ContainerList(ctx, types.ContainerListOptions{Filters: filters, All: true})
	if err != nil {
		panic(err)
	}
	if len(list) < 1 {
		return
	}
	fmt.Println("removing the container")
	err = StopAndRemoveContainer(cli, ctx, list[0].ID)
	if err != nil {
		panic(err)
	}
	fmt.Println("container removed")

}

func StopAndRemoveContainer(cli *client.Client, ctx context.Context, containerID string) error {
	err = cli.ContainerStop(ctx, containerID, nil)
	if err != nil {
		return err
	}

	err = cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{})
	if err != nil {
		return err
	}

	return nil
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
