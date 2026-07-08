package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var errWatchConfigNotFound = errors.New("watch config not found")

func (c *Client) Watch(ctx context.Context, onChange func(ConfigChangeEvent)) error {
	c.mu.Lock()
	c.onChange = onChange
	c.mu.Unlock()

	go c.watchLoop(ctx)

	return nil
}

func (c *Client) watchLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		default:
		}

		data, changed, err := c.watchOnce(ctx)
		if err != nil {
			if errors.Is(err, errWatchConfigNotFound) {
				return
			}
			time.Sleep(2 * time.Second)
			continue
		}

		if !changed {
			continue
		}

		c.applyUpdate(*data)
	}
}

func (c *Client) watchOnce(ctx context.Context) (*configData, bool, error) {
	c.mu.RLock()
	lastRevision := c.revision
	c.mu.RUnlock()

	requestURL := fmt.Sprintf(
		"%s/api/watch?app=%s&env=%s&last_revision=%d",
		c.serverAddr,
		url.QueryEscape(c.app),
		url.QueryEscape(c.env),
		lastRevision,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, false, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return nil, false, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, false, errWatchConfigNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("watch failed: status %d", resp.StatusCode)
	}

	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, false, err
	}
	if apiResp.Code != 0 {
		if strings.Contains(strings.ToLower(apiResp.Message), "config not found") {
			return nil, false, errWatchConfigNotFound
		}
		return nil, false, fmt.Errorf("watch failed: %s", apiResp.Message)
	}

	var data configData
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		return nil, false, err
	}

	return &data, true, nil
}

func (c *Client) applyUpdate(data configData) {
	c.mu.Lock()
	prev := cloneConfigs(c.configs)
	c.configs = cloneConfigs(data.Configs)
	c.revision = data.Revision
	c.version = data.Version
	onChange := c.onChange
	c.mu.Unlock()

	if onChange != nil {
		onChange(ConfigChangeEvent{
			App:        data.App,
			Env:        data.Env,
			Version:    data.Version,
			Revision:   data.Revision,
			Configs:    cloneConfigs(data.Configs),
			PrevConfig: prev,
		})
	}
}
