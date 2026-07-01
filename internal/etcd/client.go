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

func (c *Client) Get(ctx context.Context, key string) (string, bool, int64, error) {
	resp, err := c.cli.Get(ctx, key)
	if err != nil {
		return "", false, 0, err
	}

	if len(resp.Kvs) == 0 {
		return "", false, resp.Header.Revision, nil
	}

	return string(resp.Kvs[0].Value), true, resp.Header.Revision, nil
}

func (c *Client) Put(ctx context.Context, key, value string) (int64, error) {
	resp, err := c.cli.Put(ctx, key, value)
	if err != nil {
		return 0, err
	}

	return resp.Header.Revision, nil
}
