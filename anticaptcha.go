package anticaptcha

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	hcpatcha "github.com/bogdanfinn/anticaptcha/hcaptcha"
	"github.com/bogdanfinn/anticaptcha/image"
	"github.com/bogdanfinn/anticaptcha/recaptcha"
	"net/http"
	"net/url"
	"time"
)

var (
	baseURL       = &url.URL{Host: "api.anti-captcha.com", Scheme: "https", Path: "/"}
	checkInterval = 2 * time.Second
)

type Client interface {
	SolveRecaptcha(websiteURL string, recaptchaKey string, proxy string, userAgent string) (string, error)
	SolveHcaptcha(websiteURL string, siteKey string, proxy string, userAgent string) (string, error)
	GetBalance() (float64, error)
	SolveImage(imgString string) (string, error)
}

type client struct {
	apiKey          string
	timeoutInterval time.Duration
	recaptchaClient recaptcha.RecaptchaClient
	imageClient     image.ImageClient
	hcaptchaClient  hcpatcha.HcaptchaClient
}

func NewClient(apiKey string, timeoutInterval time.Duration) Client {
	return &client{
		apiKey:          apiKey,
		timeoutInterval: timeoutInterval,
		recaptchaClient: recaptcha.NewRecaptchaClient(apiKey, baseURL),
		imageClient:     image.NewImageClient(apiKey, baseURL),
		hcaptchaClient:  hcpatcha.NewHcaptchaClient(apiKey, baseURL),
	}
}

func (c *client) getTaskResult(taskID float64) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"clientKey": c.apiKey,
		"taskId":    taskID,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	u := baseURL.ResolveReference(&url.URL{Path: "/getTaskResult"})
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&responseBody)
	return responseBody, nil
}

func (c *client) SolveRecaptcha(websiteURL string, recaptchaKey string, proxy string, userAgent string) (string, error) {
	var taskID float64
	var err error

	if proxy == "" {
		taskID, err = c.recaptchaClient.CreateTaskRecaptchaProxyless(websiteURL, recaptchaKey)
	} else {
		taskID, err = c.recaptchaClient.CreateTaskRecaptcha(websiteURL, recaptchaKey, proxy, userAgent)
	}

	if err != nil {
		return "", err
	}

	check := time.NewTicker(10 * time.Second)
	timeout := time.NewTimer(c.timeoutInterval)

	for {
		select {
		case <-check.C:
			response, err := c.getTaskResult(taskID)
			if err != nil {
				return "", err
			}
			if response["status"] == "ready" {
				return response["solution"].(map[string]interface{})["gRecaptchaResponse"].(string), nil
			}
			check = time.NewTicker(checkInterval)
		case <-timeout.C:
			return "", errors.New("antiCaptcha check result timeout")
		}
	}
}

func (c *client) SolveHcaptcha(websiteURL string, siteKey string, proxy string, userAgent string) (string, error) {
	var taskID float64
	var err error

	if proxy == "" {
		taskID, err = c.hcaptchaClient.CreateTaskHcaptchaProxyless(websiteURL, siteKey)
	} else {
		taskID, err = c.hcaptchaClient.CreateTaskHcaptcha(websiteURL, siteKey, proxy, userAgent)
	}

	if err != nil {
		return "", err
	}

	check := time.NewTicker(10 * time.Second)
	timeout := time.NewTimer(c.timeoutInterval)

	for {
		select {
		case <-check.C:
			response, err := c.getTaskResult(taskID)
			if err != nil {
				return "", err
			}
			if response["status"] == "ready" {
				return response["solution"].(map[string]interface{})["gRecaptchaResponse"].(string), nil
			}
			check = time.NewTicker(checkInterval)
		case <-timeout.C:
			return "", errors.New("antiCaptcha check result timeout")
		}
	}
}

func (c *client) SolveImage(imgString string) (string, error) {
	// Create the task on anti-captcha api and get the task_id
	taskID, err := c.imageClient.CreateTaskImage(imgString)
	if err != nil {
		return "", err
	}

	// Check if the result is ready, if not loop until it is
	response, err := c.getTaskResult(taskID)
	if err != nil {
		return "", err
	}
	for {
		if response["status"] == "processing" {
			time.Sleep(checkInterval)
			response, err = c.getTaskResult(taskID)
			if err != nil {
				return "", err
			}
		} else {
			break
		}
	}

	errorDescription, ok := response["errorDescription"]
	if ok {
		return "", fmt.Errorf("server returned: %s", errorDescription)
	}

	solution, ok := response["solution"]
	if !ok {
		return "", errors.New("failed to get a response")
	}

	return solution.(map[string]interface{})["text"].(string), nil
}

func (c *client) GetBalance() (float64, error) {
	body := map[string]interface{}{
		"clientKey": c.apiKey}

	b, err := json.Marshal(body)
	if err != nil {
		return -1, err
	}

	u := baseURL.ResolveReference(&url.URL{Path: "/getBalance"})
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	responseBody := make(map[string]interface{})
	json.NewDecoder(resp.Body).Decode(&responseBody)

	errorDescription, ok := responseBody["errorDescription"]
	if ok {
		return 0, fmt.Errorf("server returned: %s", errorDescription)
	}

	balance, ok := responseBody["balance"]
	if !ok {
		return 0, errors.New("failed to get a response")
	}

	return balance.(float64), nil
}
