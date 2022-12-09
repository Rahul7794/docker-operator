package docker

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"docker-operator/config"
	"docker-operator/src/docker"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type event struct {
	Image              string            `json:"image"`
	Tag                string            `json:"tag"`
	RequestTime        time.Time         `json:"request_time"`
	Params             []string          `json:"params"`
	Method             string            `json:"method"`
	ResponseTime       time.Time         `json:"response_time"`
	ImageExistsInLocal bool              `json:"image_exists_in_local"`
	Headers            map[string]string `json:"headers,omitempty"`
	Error              string            `json:"error,omitempty"`
	ContentMD5         []byte            `json:"content_md5,omitempty"`
	Content            string            `json:"content,omitempty"`
}

type imageTag struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func RunContainerGet(dockerService docker.ServiceInterface) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var params []string
		ctx := c.Context()
		registry := config.DefaultConfig.GetString("REGISTRY")
		query := ctx.QueryArgs().String()
		if query != "" {
			params = append(params, query)
		}
		imageName := c.Params("image_name")
		tag := c.Params("tag")
		image := fmt.Sprintf("%s/%s:%s", registry, imageName, tag)
		if tag == "latest" {
			originalTag, _ := getOriginalTag(registry, imageName)
			tag = originalTag
		}
		exists := dockerService.ImageExists(image, context.Background())
		event := event{
			Image:              imageName,
			Tag:                tag,
			RequestTime:        time.Now(),
			Params:             params,
			Method:             c.Method(),
			ImageExistsInLocal: exists,
		}
		out, header, err := dockerService.RunContainer(image, params, context.Background())
		if err != nil {
			logRequestAndResponse(event, err.Error(), nil)
			if err == docker.NotFoundError {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": true,
					"msg":   err.Error(),
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}
		err = c.Send(out)
		for key, value := range header.Header {
			c.Set(key, value)
		}
		event.ContentMD5 = getMD5Hash(out)
		event.Content = firstNCharacter(string(out))
		logRequestAndResponse(event, "", header.Header)
		return err
	}
}

func RunContainerPost(dockerService docker.ServiceInterface) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		requestBody := c.Body()
		registry := config.DefaultConfig.GetString("REGISTRY")
		var params []string
		if string(requestBody) != "" {
			params = append(params, fmt.Sprintf("POST_DATA=%s", string(requestBody)))
		}
		imageName := c.Params("image_name")
		tag := c.Params("tag")
		image := fmt.Sprintf("%s/%s:%s", registry, imageName, tag)
		exists := dockerService.ImageExists(image, context.Background())
		if tag == "latest" {
			originalTag, _ := getOriginalTag(registry, imageName)
			tag = originalTag
		}
		event := event{
			Image:              imageName,
			RequestTime:        time.Now(),
			Params:             params,
			Method:             c.Method(),
			ImageExistsInLocal: exists,
		}
		out, header, err := dockerService.RunContainerPost(image, params, context.Background())
		if err != nil {
			logRequestAndResponse(event, err.Error(), nil)
			if err == docker.NotFoundError {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": true,
					"msg":   err.Error(),
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": true,
				"msg":   err.Error(),
			})
		}
		err = c.Send(out)
		for key, value := range header.Header {
			c.Set(key, value)
		}
		event.ContentMD5 = getMD5Hash(out)
		event.Content = firstNCharacter(string(out))
		event.Tag = tag
		logRequestAndResponse(event, "", header.Header)
		return err
	}
}

func logRequestAndResponse(event event, err string, header map[string]string) {
	event.Error = err
	event.ResponseTime = time.Now()
	event.Headers = header
	zap.S().With(
		"request", event,
	).Infow("incoming request")
}

func getMD5Hash(data []byte) []byte {
	hash := md5.New()
	hash.Write(data)
	return hash.Sum(nil)
}

func firstNCharacter(data string) string {
	n := config.DefaultConfig.GetInt("CONTENT_LENGTH")
	if len(data) < n {
		return data
	}
	return data[0:n]
}

func getOriginalTag(repository, imageName string) (string, error) {
	url := fmt.Sprintf(`https://%s/v2/%s/tags/list`, repository, imageName)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	imageTagResponse := &imageTag{}
	err = json.NewDecoder(resp.Body).Decode(imageTagResponse)
	if err != nil {
		return "", err
	}
	tags := imageTagResponse.Tags
	sort.Strings(tags)
	return tags[len(tags)-2], nil
}
