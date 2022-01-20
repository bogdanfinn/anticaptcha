package recaptcha

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

type RecaptchaClient interface {
	CreateTaskRecaptcha(websiteURL string, recaptchaKey string, proxy string, userAgent string) (float64, error)
	CreateTaskRecaptchaProxyless(websiteURL string, recaptchaKey string) (float64, error)
}

type recaptchaClient struct {
	apiKey  string
	baseUrl *url.URL
}

func NewRecaptchaClient(apiKey string, baseUrl *url.URL) RecaptchaClient {
	return &recaptchaClient{
		apiKey: apiKey,
		baseUrl: baseUrl,
	}
}

func (c *recaptchaClient) CreateTaskRecaptcha(websiteURL string, recaptchaKey string, proxy string, userAgent string) (float64, error) {
	body := map[string]interface{}{
		"clientKey": c.apiKey,
		"task": map[string]interface{}{
			"type":       "RecaptchaV2Task",
			"websiteURL": websiteURL,
			"websiteKey": recaptchaKey,
			"proxyType": "http",
			"proxyAddress": "",
			"proxyPort": "",
			"proxyLogin": "",
			"proxyPassword": "",
			"userAgent": "",
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

	if _, ok := responseBody["taskId"]; ok {
		if taskId, ok := responseBody["taskId"].(float64); ok {
			return taskId, nil
		}

		return 0, errors.New("task number of irregular format")
	}

	return 0, errors.New("task number not found in server response")
}


func (c *recaptchaClient) CreateTaskRecaptchaProxyless(websiteURL string, recaptchaKey string) (float64, error) {
	body := map[string]interface{}{
		"clientKey": c.apiKey,
		"task": map[string]interface{}{
			"type":       "RecaptchaV2TaskProxyless",
			"websiteURL": websiteURL,
			"websiteKey": recaptchaKey,
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

	if _, ok := responseBody["taskId"]; ok {
		if taskId, ok := responseBody["taskId"].(float64); ok {
			return taskId, nil
		}

		return 0, errors.New("task number of irregular format")
	}

	return 0, errors.New("task number not found in server response")
}
