// This file is part of bizfly-agent
//
// Copyright (C) 2020  BizFly Cloud
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>

package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	prol "github.com/prometheus/common/log"

	"github.com/bizflycloud/bizfly-agent/auth"
	"github.com/bizflycloud/bizfly-agent/config"
)

// Client ...
type Client struct {
	httpClient      *http.Client
	defaultEndpoint string
	token           string
	authToken       *auth.Token
}

// NewHTTPClient ...
func NewHTTPClient() *Client {
	c := &Client{
		httpClient: http.DefaultClient,
	}
	if at, _ := auth.NewToken(); at != nil {
		c.authToken = at
	}
	return c
}

// AuthToken get, save and set auth token
func (c *Client) AuthToken() (string, error) {
	if config.Config.AuthServer.DefaultEndpoint == "" {
		prol.Fatalln("Default Endpoint is required")
	}
	c.defaultEndpoint = config.Config.AuthServer.DefaultEndpoint
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/agent_tokens?agent_id=%s", c.defaultEndpoint, config.Config.Agent.ID), nil)
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}

	req.Header.Set("X-Agent-Secret", config.Config.AuthServer.Secret)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	tokenStr := string(body)

	if resp.StatusCode == http.StatusForbidden {
		prol.Fatalln("Error when get new auth token for agent")
	}
	if c.authToken != nil {
		_ = c.authToken.SaveToken(tokenStr)
	}

	c.token = tokenStr
	return tokenStr, nil
}

// Get ...
func (c *Client) Get(url string) ([]byte, error) {
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

// Do ...
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	var err error
	if c.token == "" && c.authToken != nil {
		c.token, err = c.authToken.ReadToken()
		if err != nil {
			return nil, err
		}
	}
	req.Header.Add("Authorization", "Bearer "+c.token)
	body, _ := ioutil.ReadAll(req.Body)
	req.Body = ioutil.NopCloser(bytes.NewReader(body))

	res, err := c.httpClient.Do(req)
	if err != nil || res.StatusCode != http.StatusForbidden {
		return res, err
	}

	// Maybe token expired, get new one and retry
	c.token, err = c.AuthToken()
	if err != nil {
		return nil, err
	}

	req.Body = ioutil.NopCloser(bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+c.token)
	return c.httpClient.Do(req)
}
