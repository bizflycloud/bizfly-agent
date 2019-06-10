package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const defaultMetadataEndpoint = "http://169.254.169.254/metadata"

type client struct {
	httpClient       *http.Client
	metadataEndpoint string
	authToken        string
}

func (c *client) GetAuthToken() (string, error) {
	return c.get(fmt.Sprintf("%s/agent/authtoken", c.metadataEndpoint))
}

func (c *client) get(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req.WithContext(context.Background()))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New("failed to get")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *client) Do(req *http.Request) (*http.Response, error) {
	var err error
	if c.authToken == "" {
		c.authToken, err = c.GetAuthToken()
		if err != nil {
			return nil, err
		}
	}
	req.Header.Add("Authorization", "Bearer "+c.authToken)

	return c.httpClient.Do(req)
}
