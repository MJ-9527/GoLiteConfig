package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Client struct {
	serverAddr string
	app        string
	env        string
	httpClient *http.Client

	mu       sync.RWMutex
	configs  map[string]string
	revision int64
	version  string

	onChange func(ConfigChangeEvent)
	stopCh   chan struct{}
}

func NewClient(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.ServerAddr) == "" {
		return nil, fmt.Errorf("server address is required")
	}
	if strings.TrimSpace(cfg.App) == "" {
		return nil, fmt.Errorf("app is required")
	}
	if strings.TrimSpace(cfg.Env) == "" {
		return nil, fmt.Errorf("env is required")
	}

	if cfg.PollTimeout <= 0 {
		cfg.PollTimeout = time.Second * 65
	}

	client := &Client{
		serverAddr: strings.TrimRight(cfg.ServerAddr, "/"),
		app:        cfg.App,
		env:        cfg.Env,
		httpClient: &http.Client{
			Timeout: cfg.PollTimeout,
		},
		configs: make(map[string]string),
		stopCh:  make(chan struct{}),
	}

	if cfg.AutoLoadOnStart {
		if err := client.loadInitial(context.Background()); err != nil {
			return nil, err
		}
	}

	return client, nil
}

func (c *Client) fetchCurrent(ctx context.Context) (*configData, error) {
	requestURL := fmt.Sprintf(
		"%s/api/config?app=%s&env=%s",
		c.serverAddr,
		url.QueryEscape(c.app),
		url.QueryEscape(c.env),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch current config failed: status %d", resp.StatusCode)
	}

	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	if apiResp.Code != 0 {
		return nil, fmt.Errorf("fetch current config failed: %s", apiResp.Message)
	}

	var data configData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (c *Client) loadInitial(ctx context.Context) error {
	data, err := c.fetchCurrent(ctx)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.configs = cloneConfigs(data.Configs)
	c.revision = data.Revision
	c.version = data.Version

	return nil
}

func (c *Client) Load(ctx context.Context) error {
	return c.loadInitial(ctx)
}

func (c *Client) Get(key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, ok := c.configs[key]
	if !ok {
		return "", fmt.Errorf("config key not found: %s", key)
	}

	return value, nil
}

func (c *Client) GetJSON(key string, target any) error {
	value, err := c.Get(key)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(value), target)
}

func (c *Client) Close() error {
	select {
	case <-c.stopCh:
	default:
		close(c.stopCh)
	}

	return nil
}

func cloneConfigs(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
