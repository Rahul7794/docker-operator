package tests

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"docker-operator/src/docker"
	routes "docker-operator/src/v1"

	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestExecImageGET(t *testing.T) {
	tests := []struct {
		description        string
		route              string
		method             string
		expectedError      bool
		expectedStatusCode int
		mockService        func(ms *docker.MockServiceInterface) *docker.MockServiceInterface
		expectedBody       []byte
	}{
		{
			description:        "execute the image",
			route:              "/api/exec/alpine/latest",
			method:             "GET",
			expectedError:      false,
			expectedStatusCode: http.StatusOK,
			mockService: func(ms *docker.MockServiceInterface) *docker.MockServiceInterface {
				ms.EXPECT().RunContainer(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte(`content of response`),
					&docker.Headers{Header: map[string]string{}}, nil)
				ms.EXPECT().ImageExists(gomock.Any(), gomock.Any()).Return(true)
				return ms
			},
			expectedBody: []byte(`content of response`),
		},
		{
			description:        "return 500 internal error",
			route:              "/api/exec/alpine/latest",
			method:             "GET",
			expectedError:      false,
			expectedStatusCode: http.StatusInternalServerError,
			mockService: func(ms *docker.MockServiceInterface) *docker.MockServiceInterface {
				ms.EXPECT().RunContainer(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, &docker.Headers{Header: map[string]string{}},
					fmt.Errorf("internal error"))
				ms.EXPECT().ImageExists(gomock.Any(), gomock.Any()).Return(true)
				return ms
			},
			expectedBody: []byte(`{"error":true,"msg":"internal error"}`),
		},
	}

	for _, test := range tests {

		t.Run(test.description, func(t *testing.T) {
			app := fiber.New()
			ctrl := gomock.NewController(t)
			dockerService := test.mockService(docker.NewMockServiceInterface(ctrl))
			routes.AddRoutes(app, dockerService)
			req := httptest.NewRequest(test.method, test.route, nil)
			resp, err := app.Test(req, -1) // the -1 disables request latency
			assert.Equalf(t, test.expectedError, err != nil, test.description)
			if test.expectedError {
				return
			}
			// Check the status code is what we expect.
			assert.Equalf(t, test.expectedStatusCode, resp.StatusCode, test.description)

			// Check the response body is what we expect.
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fail()
			}

			assert.Equalf(t, test.expectedBody, body, test.description)
		})
	}
}

func TestExecImagePOST(t *testing.T) {
	tests := []struct {
		description        string
		route              string
		method             string
		expectedError      bool
		requestBody        io.Reader
		expectedStatusCode int
		mockService        func(ms *docker.MockServiceInterface) *docker.MockServiceInterface
		expectedBody       []byte
	}{
		{
			description:        "execute the image",
			route:              "/api/exec/alpine/latest",
			method:             "POST",
			expectedError:      false,
			requestBody:        strings.NewReader(`{"data": ["00000X71080", "json"]}`),
			expectedStatusCode: http.StatusOK,
			mockService: func(ms *docker.MockServiceInterface) *docker.MockServiceInterface {
				ms.EXPECT().RunContainer(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte(`content of response`),
					&docker.Headers{Header: map[string]string{}}, nil).AnyTimes()
				ms.EXPECT().ImageExists(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				ms.EXPECT().RunContainerPost(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte(`content of response`),
					&docker.Headers{Header: map[string]string{}}, nil).AnyTimes()
				return ms
			},
			expectedBody: []byte(`content of response`),
		},
		{
			description:        "return 500 internal error",
			route:              "/api/exec/alpine/latest",
			method:             "POST",
			expectedError:      false,
			requestBody:        strings.NewReader(`{"data": ["00000X71080", "json"]}`),
			expectedStatusCode: http.StatusInternalServerError,
			mockService: func(ms *docker.MockServiceInterface) *docker.MockServiceInterface {
				ms.EXPECT().RunContainer(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, &docker.Headers{Header: map[string]string{}},
					fmt.Errorf("internal error")).AnyTimes()
				ms.EXPECT().ImageExists(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				ms.EXPECT().RunContainerPost(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, &docker.Headers{Header: map[string]string{}},
					fmt.Errorf("internal error")).AnyTimes()
				return ms
			},
			expectedBody: []byte(`{"error":true,"msg":"internal error"}`),
		},
	}

	for _, test := range tests {

		t.Run(test.description, func(t *testing.T) {
			app := fiber.New()
			ctrl := gomock.NewController(t)
			dockerService := test.mockService(docker.NewMockServiceInterface(ctrl))
			routes.AddRoutes(app, dockerService)
			req := httptest.NewRequest(test.method, test.route, test.requestBody)
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req, -1) // the -1 disables request latency
			assert.Equalf(t, test.expectedError, err != nil, test.description)
			if test.expectedError {
				return
			}
			// Check the status code is what we expect.
			assert.Equalf(t, test.expectedStatusCode, resp.StatusCode, test.description)

			// Check the response body is what we expect.
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fail()
			}

			fmt.Println(string(body))

			assert.Equalf(t, test.expectedBody, body, test.description)
		})
	}
}
