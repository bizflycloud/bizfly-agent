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
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

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
		httpClient: &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
	if at, err := auth.NewToken(); at != nil {
		c.authToken = at
	} else {
		prol.Warnf("Token cached is disabled, auth.NewToken() failed: %v", err)
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
		prol.Error("Error reading request. ", err)
		return "", err
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
		prol.Error("Error when get new auth token for agent")
	}
	if c.authToken != nil {
		_ = c.authToken.SaveToken(tokenStr)
	}

	c.token = tokenStr
	return tokenStr, nil
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
