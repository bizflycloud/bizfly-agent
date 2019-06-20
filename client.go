package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const defaultMetadataEndpoint = "http://169.254.169.254"

type client struct {
	httpClient       *http.Client
	metadataEndpoint string
	authToken        string
}

func newHTTPClient() *client {
	return &client{
		httpClient:       http.DefaultClient,
		metadataEndpoint: defaultMetadataEndpoint,
	}
}

func (c *client) AuthToken() (string, error) {
	resp, err := c.get(fmt.Sprintf("%s/metadata/agent/authtoken", c.metadataEndpoint))
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func (c *client) get(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req.WithContext(context.Background()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New("failed to get")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (c *client) Do(req *http.Request) (*http.Response, error) {
	var err error
	if c.authToken == "" {
		c.authToken, err = c.AuthToken()
		if err != nil {
			return nil, err
		}
	}
	req.Header.Add("Authorization", "Bearer "+c.authToken)

	res, err := c.httpClient.Do(req)
	if err != nil || res.StatusCode != http.StatusForbidden {
		return res, err
	}

	// Maybe token expired, get new one and retry
	c.authToken, err = c.AuthToken()
	if err != nil {
		return nil, err
	}

	return c.httpClient.Do(req)
}
