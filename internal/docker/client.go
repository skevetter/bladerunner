package docker

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

type Client struct {
	cli *client.Client
}

func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Client{cli: cli}, nil
}

func (c *Client) StartRunnerContainer(ctx context.Context, image string, env []string, name string) (string, error) {
	logrus.Infof("Starting runner container %s using image %s", name, image)

	// Pull image if not exists (simplified)
	// In production, might want to explicitly pull or check

	resp, err := c.cli.ContainerCreate(ctx, &container.Config{
		Image: image,
		Env:   env,
		Cmd:   []string{"bladerunner", "runner", "start"},
	}, &container.HostConfig{
		AutoRemove: true, // Remove container when it exits
		Binds: []string{
			"/var/run/docker.sock:/var/run/docker.sock", // Docker-in-Docker support
		},
		Privileged: true, // Required for Docker-in-Docker
	}, nil, nil, name)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	if err := c.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	return resp.ID, nil
}

func (c *Client) StopContainer(ctx context.Context, containerID string) error {
	logrus.Infof("Stopping container %s", containerID)
	// Send SIGTERM first
	timeout := 30 // seconds
	return c.cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})
}

func (c *Client) ListRunnerContainers(ctx context.Context) ([]string, error) {
	containers, err := c.cli.ContainerList(ctx, types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return nil, err
	}

	var ids []string
	for _, c := range containers {
		// Filter by label or name if needed
		ids = append(ids, c.ID)
	}
	return ids, nil
}

func (c *Client) PullImage(ctx context.Context, imageName string) error {
	reader, err := c.cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	io.Copy(os.Stdout, reader) // Stream output
	return nil
}
