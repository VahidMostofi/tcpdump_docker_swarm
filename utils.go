package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

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

func getSingleContainerBySingleFilter(cli *client.Client, ctx context.Context, filterKey, filterValue string) (*types.Container, error) {
	containerListFilters := filters.NewArgs()
	containerListFilters.Add(filterKey, filterValue)

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true, Filters: containerListFilters})
	if err != nil {
		return nil, err
	}
	if len(containers) == 0 {
		return nil, nil
	}
	return &containers[0], nil
}

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
