package sdk

import (
	"context"
	"encoding/json"
	"time"
)

type Config struct {
	ServerAddr       string
	App              string
	Env              string
	PollTimeout      time.Duration
	AutoLoadOnStart  bool
	AutoWatchOnStart bool
}

type ConfigProvider interface {
	Get(key string) (string, error)
	GetJSON(key string, target any) error
	Watch(ctx context.Context, onChange func(ConfigChangeEvent)) error
	Close() error
}

type ConfigChangeEvent struct {
	App        string
	Env        string
	Version    string
	Revision   int64
	Configs    map[string]string
	PrevConfig map[string]string
}

type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type configData struct {
	App      string            `json:"app"`
	Env      string            `json:"env"`
	Version  string            `json:"version"`
	Revision int64             `json:"revision"`
	Configs  map[string]string `json:"configs"`
}
