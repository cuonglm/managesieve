// Copyright 2020 Cuong Manh Le. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package managesieve implements ManageSieve client protocol.
package managesieve

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// Client manages communication with managesieve server.
type Client struct {
	conn    net.Conn
	scanner *bufio.Scanner

	addr string
}

// ClientOption configures Client.
type ClientOption func(*Client) error

// WithServerAddress sets the managesieve server address of Client.
func WithServerAddress(addr string) ClientOption {
	return func(c *Client) error {
		c.addr = addr
		return nil
	}
}

// WithConn sets the underlying connection used by Client, for testing purpose.
func WithConn(conn net.Conn) ClientOption {
	return func(c *Client) error {
		c.conn = conn
		return nil
	}
}

// NewClient returns a new managesieve client.
func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{addr: "localhost:4190"}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	conn, err := net.DialTimeout("tcp", c.addr, 5*time.Second)
	if err != nil {
		return nil, err
	}
	c.conn = conn
	c.scanner = bufio.NewScanner(c.conn)
	_ = c.readResponse()

	return c, nil
}

func (c *Client) readResponse() error {
	var err error
	for c.scanner.Scan() {
		line := strings.ToUpper(c.scanner.Text())
		if strings.HasPrefix(line, "OK") {
			err = c.scanner.Err()
			break
		}
		if strings.HasPrefix(line, "NO") {
			err = errors.New(line[2:])
			break
		}
		if strings.HasPrefix(line, "BYE") {
			err = errors.New(line[3:])
			break
		}
	}
	return err
}

func (c *Client) runCmd(cmd string, args ...string) error {
	for i, arg := range args {
		args[i] = strconv.Quote(arg)
	}
	_, _ = fmt.Fprint(c.conn, cmd, " ", strings.Join(args, " "), "\r\n")

	return c.readResponse()
}

// Login authenticates with managesieve server with given username and password,
// using PLAIN auth.
func (c *Client) Login(user, pass string) error {
	auth := base64.StdEncoding.EncodeToString([]byte("\x00" + user + "\x00" + pass))
	return c.runCmd("AUTHENTICATE", "PLAIN", auth)
}

// GetScript gets sieve script by name.
func (c *Client) GetScript(name string) error {
	return c.runCmd("GETSCRIPT", name)
}

// PutScript replace a sieve script with new content.
func (c *Client) PutScript(name string, content string) error {
	content = fmt.Sprintf("{%d+}\r\n%s", len(content), content)
	return c.runCmd("PUTSCRIPT", name, content)
}

// SetActive marks the sieve script active.
func (c *Client) SetActive(name string) error {
	return c.runCmd("SETACTIVE", name)
}

// DeleteScript deletes a sieve script by name.
func (c *Client) DeleteScript(name string) error {
	return c.runCmd("DELETESCRIPT", name)
}
