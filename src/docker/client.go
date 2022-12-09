package docker

import (
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

type Client struct {
	Client ClientInterface
}

//go:generate mockgen -source=src/docker/client.go -package docker -destination src/docker/client_mock.go
type ClientInterface interface {
	ImagePull(ctx context.Context, refStr string, options types.ImagePullOptions) (io.ReadCloser, error)
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *specs.Platform, containerName string) (container.ContainerCreateCreatedBody, error)
	ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error
	ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.ContainerWaitOKBody, <-chan error)
	ContainerLogs(ctx context.Context, container string, options types.ContainerLogsOptions) (io.ReadCloser, error)
	ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error
	ImageInspectWithRaw(ctx context.Context, imageID string) (types.ImageInspect, []byte, error)
}

func (s *Client) ImagePull(ctx context.Context, refStr string, options types.ImagePullOptions) (io.ReadCloser, error) {
	return s.Client.ImagePull(ctx, refStr, options)
}

func (s *Client) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *specs.Platform, containerName string) (container.ContainerCreateCreatedBody, error) {
	return s.Client.ContainerCreate(ctx, config, hostConfig, networkingConfig, platform, containerName)
}

func (s *Client) ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error {
	return s.Client.ContainerStart(ctx, containerID, options)
}

func (s *Client) ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.ContainerWaitOKBody, <-chan error) {
	return s.Client.ContainerWait(ctx, containerID, condition)
}

func (s *Client) ContainerLogs(ctx context.Context, container string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
	return s.Client.ContainerLogs(ctx, container, options)
}

func (s *Client) ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error {
	return s.Client.ContainerRemove(ctx, containerID, options)
}

func (s *Client) ImageInspectWithRaw(ctx context.Context, imageID string) (types.ImageInspect, []byte, error){
	return s.Client.ImageInspectWithRaw(ctx, imageID)
}

func NewClient() (ClientInterface, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("could not create docker cli client with: %w", err)
	}
	return &Client{
		Client: cli,
	}, nil
}
