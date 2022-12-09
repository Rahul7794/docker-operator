package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"go.uber.org/zap"
)

var NotFoundError = fmt.Errorf("image not found")
var ContainerRunError = fmt.Errorf("error occured while running the image")

type Service struct {
	Client *Client
}

type Headers struct {
	Header map[string]string
}

//go:generate mockgen -source=src/docker/service.go -package docker -destination src/docker/service_mock.go
type ServiceInterface interface {
	RunContainer(image string, params []string, ctx context.Context) ([]byte, *Headers, error)
	RunContainerPost(image string, reqBody []string, ctx context.Context) ([]byte, *Headers, error)
	ImageExists(image string, ctx context.Context) bool
}

func (s *Service) RunContainer(image string, params []string, ctx context.Context) ([]byte, *Headers, error) {
	reader, err := s.Client.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		zap.S().Error(NotFoundError)
		return nil, nil, fmt.Errorf("%w, %v", NotFoundError, err)
	}
	defer reader.Close()
	io.Copy(ioutil.Discard, reader)
	containerConfig := &container.Config{
		Image: image,
		Cmd:   params,
		Tty:   false,
	}
	return runImage(containerConfig, ctx, s)
}

func (s *Service) RunContainerPost(image string, reqBody []string, ctx context.Context) ([]byte, *Headers, error) {
	reader, err := s.Client.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		zap.S().Error(NotFoundError)
		return nil, nil, fmt.Errorf("%w, %v", NotFoundError, err)
	}
	defer reader.Close()
	io.Copy(ioutil.Discard, reader)
	containerConfig := &container.Config{
		Image: image,
		Env:   reqBody,
		Tty:   false,
	}
	return runImage(containerConfig, ctx, s)
}
func runImage(config *container.Config, ctx context.Context, s *Service) ([]byte, *Headers, error) {
	resp, err := s.Client.ContainerCreate(ctx, config, nil, nil,nil, "")
	if err != nil {
		errMessage := fmt.Errorf("could not create a new container for image %s because: %w", config.Image, err)
		zap.S().Error(errMessage.Error())
		return nil, nil, errMessage
	}

	if err := s.Client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		errMessage := fmt.Errorf("could not start container with: %w", err)
		zap.S().Error(errMessage.Error())
		return nil, nil, errMessage
	}

	statusCh, errCh := s.Client.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			errMessage := fmt.Errorf("cannot wait for container to complete with: %w", err)
			zap.S().Error(errMessage.Error())
			return nil, nil, errMessage
		}
	case <-statusCh:
	}

	out, err := s.Client.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStderr: true,
		ShowStdout: true,
		Follow:     true})
	if err != nil {
		errMessage := fmt.Errorf("cannot get container logs with: %w", err)
		zap.S().Error(errMessage.Error())
		return nil, nil, errMessage
	}

	if err = s.Client.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{}); err != nil {
		errMessage := fmt.Errorf("cannot stop containers: %w", err)
		zap.S().Error(errMessage.Error())
		return nil, nil, errMessage
	}

	buffer := &bytes.Buffer{}
	errorWriter := &bytes.Buffer{}
	_, err = stdcopy.StdCopy(buffer, errorWriter, out)
	if err != nil {
		errMessage := fmt.Errorf("cannot read logs from the container: %w", err)
		zap.S().Error(errMessage.Error())
		return nil, nil, errMessage
	}
	errorString := errorWriter.String()
	if len(errorString) > 0 {
		zap.L().Error(errorString)
		return nil, nil, ContainerRunError
	}
	return processContainerLogs(buffer.String())
}

func processContainerLogs(out string) ([]byte, *Headers, error) {
	logsSplit := strings.SplitN(out, "\n\n", 2)
	headerMap, hasContentType := processHeader(logsSplit[0])
	if !hasContentType {
		errMessage := fmt.Errorf("does not contain content type in logs")
		zap.S().Error(errMessage.Error())
		return nil, nil, errMessage
	}
	return []byte(logsSplit[1]), &Headers{Header: headerMap}, nil
}

func processHeader(str string) (map[string]string, bool) {
	var hasContentType bool
	headerMap := make(map[string]string)
	headers := strings.Split(str, "\n")
	for _, header := range headers {
		if strings.HasPrefix(strings.ToLower(header), "content-type") {
			hasContentType = true
		}
		splitHeader := strings.SplitN(header, ":", 2)
		if len(splitHeader) > 1 {
			headerMap[splitHeader[0]] = strings.TrimSpace(splitHeader[1])
		}
	}
	return headerMap, hasContentType
}

func (s *Service) ImageExists(image string, ctx context.Context) bool {
	_, _, err := s.Client.ImageInspectWithRaw(ctx, image)
	if err != nil {
		return false
	}
	return true
}


func NewService(clientInterface ClientInterface) ServiceInterface {
	return &Service{
		Client: &Client{clientInterface},
	}
}

