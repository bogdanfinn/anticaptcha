package image

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type ImageClient interface {
	CreateTaskImage(imgString string) (float64, error)
}

type imageClient struct {
	apiKey  string
	baseUrl *url.URL
}

func NewImageClient(apiKey string, baseUrl *url.URL) ImageClient {
	return &imageClient{
		apiKey: apiKey,
		baseUrl: baseUrl,
	}
}

func (c *imageClient) CreateTaskImage(imgString string) (float64, error) {
	body := map[string]interface{}{
		"clientKey": c.apiKey,
		"task": map[string]interface{}{
			"type": "ImageToTextTask",
			"body": imgString,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}

	u := c.baseUrl.ResolveReference(&url.URL{Path: "/createTask"})
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	responseBody := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&responseBody)

	errorDescription, ok := responseBody["errorDescription"]
	if ok {
		return 0, fmt.Errorf("server returned: %s", errorDescription)
	}

	taskId, ok := responseBody["taskId"]
	if !ok {
		return 0, errors.New("failed to get a response")
	}

	return taskId.(float64), nil
}
