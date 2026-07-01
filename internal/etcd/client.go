package etcd

import (
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Client struct {
	cli *clientv3.Client
}

func NewClient(endpoints []string) (*Client, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return &Client{cli: cli}, nil
}

func (c *Client) Close() error {
	if c == nil || c.cli == nil {
		return nil
	}

	return c.cli.Close()
}

func (c *Client) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	_, err := c.cli.Status(pingCtx, c.cli.Endpoints()[0])
	return err
}

func (c *Client) Raw() *clientv3.Client {
	return c.cli
}
