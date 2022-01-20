package hcpatcha

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

type HcaptchaClient interface {
	CreateTaskHcaptchaProxyless(websiteURL string, siteKey string) (float64, error)
	CreateTaskHcaptcha(websiteURL string, siteKey string, proxy string, userAgent string) (float64, error)
}

type hcaptchaClient struct {
	apiKey  string
	baseUrl *url.URL
}

func NewHcaptchaClient(apiKey string, baseUrl *url.URL) HcaptchaClient {
	return &hcaptchaClient{
		apiKey:  apiKey,
		baseUrl: baseUrl,
	}
}

func (c *hcaptchaClient) CreateTaskHcaptchaProxyless(websiteURL string, siteKey string) (float64, error) {
	body := map[string]interface{}{
		"clientKey": c.apiKey,
		"task": map[string]interface{}{
			"type":       "HCaptchaTask",
			"websiteURL": websiteURL,
			"websiteKey": siteKey,
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

func (c *hcaptchaClient) CreateTaskHcaptcha(websiteURL string, siteKey string, proxy string, userAgent string) (float64, error) {
	body := map[string]interface{}{
		"clientKey": c.apiKey,
		"task": map[string]interface{}{
			"type":       "HCaptchaTask",
			"websiteURL": websiteURL,
			"websiteKey": siteKey,
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
