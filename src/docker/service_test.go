package docker

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/golang/mock/gomock"
)

func TestService_RunContainer(t *testing.T) {
	tests := []struct {
		name  string
		mock  func(clientInterface *MockClientInterface, image string) *MockClientInterface
		err   error
		image string
	}{
		{
			name: "image not available error",
			mock: func(mc *MockClientInterface, image string) *MockClientInterface {
				mc.EXPECT().ImagePull(context.Background(), image, types.ImagePullOptions{}).Return(nil, NotFoundError)
				return mc
			},
			err:   fmt.Errorf("image not found, image not found"),
			image: "alpine",
		},
		{
			name: "image being pulled but cant create container",
			mock: func(mc *MockClientInterface, image string) *MockClientInterface {
				containerWaitChan := make(chan container.ContainerWaitOKBody)
				errorChan := make(chan error)
				defer close(containerWaitChan)
				defer close(errorChan)
				mc.EXPECT().ImagePull(context.Background(), image, types.ImagePullOptions{}).Return(stringToIOReader("pulled image successfully \n"), nil)
				mc.EXPECT().ContainerCreate(context.Background(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(container.ContainerCreateCreatedBody{}, fmt.Errorf(`/bin/sh executable not found`))
				return mc

			},
			image: "alpine",
			err:   fmt.Errorf(`could not create a new container for image alpine because: /bin/sh executable not found`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockInterface := NewMockClientInterface(ctrl)
			client := tt.mock(mockInterface, tt.image)
			service := NewService(client)
			_, _, err := service.RunContainer(tt.image, []string{"json"}, context.Background())
			if err != nil {
				if !reflect.DeepEqual(err.Error(), tt.err.Error()) {
					t.Errorf("RunContainer() gotError = %v, want = %v", err, tt.err)
				}
				return
			}
		})
	}
}

func TestProcessContainerLogs(t *testing.T) {
	tests := []struct {
		name        string
		logs        func() string
		wantContent string
		wantHeader  func() *Headers
		err         error
	}{
		{
			name: "successfully process container logs",
			logs: func() string {
				return `Content-Type: text/html
Content-Length: 149

<html><head><title>Available formats</title></head>
<body><h1>Available formats</h1>
<ul><li>html</li><li>json</li><li>plain</li></ul>
</body></html>`
			},
			wantHeader: func() *Headers {
				headerMap := make(map[string]string)
				headerMap["Content-Length"] = "149"
				headerMap["Content-Type"] = "text/html"
				return &Headers{
					headerMap,
				}
			},
			wantContent: `<html><head><title>Available formats</title></head>
<body><h1>Available formats</h1>
<ul><li>html</li><li>json</li><li>plain</li></ul>
</body></html>`,
			err: nil,
		},
		{
			name: "successfully process container logs with unusual header",
			logs: func() string {
				return `Content-Type: text/html
Content-Length: 74
Forwarded: by=_hidden;host:dev-exec.faas.it;proto=https
Unusual-Header: 945

{"count": 3, "name": "images", "content-types": ["json", "html", "plain"]}`
			},
			wantHeader: func() *Headers {
				headerMap := make(map[string]string)
				headerMap["Content-Length"] = "74"
				headerMap["Forwarded"] = "by=_hidden;host:dev-exec.faas.it;proto=https"
				headerMap["Unusual-Header"] = "945"
				headerMap["Content-Type"] = "text/html"
				return &Headers{
					headerMap,
				}
			},
			wantContent: `{"count": 3, "name": "images", "content-types": ["json", "html", "plain"]}`,
			err: nil,
		},
		{
			name: "return error when no content type returned in logs",
			logs: func() string {
				return `hello`
			},
			wantHeader: func() *Headers {
				return nil
			},
			wantContent: ``,
			err: fmt.Errorf("does not contain content type in logs"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotContent, gotHeader, err := processContainerLogs(tt.logs())
			if err != nil {
				if !reflect.DeepEqual(err.Error(), tt.err.Error()) {
					t.Errorf("processContainerLogs() gotError = %v, want = %v", err, tt.err)
				}
				return
			}
			if !reflect.DeepEqual(string(gotContent), tt.wantContent) {
				t.Errorf("processContainerLogs() gotContent = %v, want = %v", string(gotContent), tt.wantContent)
			}
			if !reflect.DeepEqual(gotHeader, tt.wantHeader()) {
				t.Errorf("processContainerLogs() gotHeader = %v, want = %v", gotHeader, tt.wantHeader())
			}
		})
	}
}

func stringToIOReader(str string) io.ReadCloser {
	return ioutil.NopCloser(strings.NewReader(str))
}
