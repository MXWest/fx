package api

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/jhoonb/archivex"

	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func Build(name string, dir string) {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	tar := new(archivex.TarFile)
	tar.Create(dir)
	tar.AddAll(dir, false)
	tar.Close()
	dockerBuildContext, buildContextErr := os.Open(dir + ".tar")
	if buildContextErr != nil {
		panic(buildContextErr)
	}
	defer dockerBuildContext.Close()

	buildOptions := types.ImageBuildOptions{
		Dockerfile: "Dockerfile", // optional, is the default
		Tags:       []string{name},
	}
	buildResponse, buildErr := cli.ImageBuild(context.Background(), dockerBuildContext, buildOptions)
	if buildErr != nil {
		panic(buildErr)
	}
	log.Println("build ", buildResponse, buildErr)

	response, err := ioutil.ReadAll(buildResponse.Body)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	fmt.Println(string(response))
}

func Deploy(name string, dir string, port string) {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	imageName := name
	containerConfig := &container.Config{
		Image: imageName,
		ExposedPorts: nat.PortSet{
			"3000/tcp": struct{}{},
		},
	}
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"3000/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: port,
				},
			},
		},
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, "")
	if err != nil {
		panic(err)
	}
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
	fmt.Println(resp.ID)
}

func Stop(containerID string) (err error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	timeout := time.Duration(1) * time.Second
	err = cli.ContainerStop(context.Background(), containerID, &timeout)
	return err
}
