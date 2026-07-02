package service

import (
	"GoLiteConfig/internal/etcd"
	"GoLiteConfig/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ConfigService struct {
	etcd *etcd.Client
}

func NewConfigService(etcdClient *etcd.Client) *ConfigService {
	return &ConfigService{etcd: etcdClient}
}

func (s *ConfigService) Publish(ctx context.Context, req model.PublishConfigRequest) (*model.PublishConfigResponse, error) {
	// 校验参数
	if req.App == "" {
		return nil, fmt.Errorf("app is required")
	}
	if req.Env == "" {
		return nil, fmt.Errorf("env is required")
	}
	if len(req.Configs) == 0 {
		return nil, fmt.Errorf("configs is required")
	}

	// 拼写路径
	baseKey := fmt.Sprintf("/config/%s/%s", req.App, req.Env)

	// 读取version_counter
	counterKey := baseKey + "/version_counter"
	counterValue, exists, _, err := s.etcd.Get(ctx, counterKey)
	if err != nil {
		return nil, err
	}

	// 生成新版本号
	nextNumber := 1
	if exists {
		currentNumber, err := strconv.Atoi(counterValue)
		if err != nil {
			return nil, fmt.Errorf("invalid version_counter: %w", err)
		}
		nextNumber = currentNumber + 1
	}

	version := fmt.Sprintf("v%d", nextNumber)

	// 写config
	type parsedConfig struct {
		group string
		key   string
		value string
	}

	parsedConfigs := make([]parsedConfig, 0, len(req.Configs))
	for configKey, configValue := range req.Configs {
		group, key, err := splitConfigKey(configKey)
		if err != nil {
			return nil, err
		}

		parsedConfigs = append(parsedConfigs, parsedConfig{
			group: group,
			key:   key,
			value: configValue,
		})
	}
	var revision int64
	for _, cfg := range parsedConfigs {
		etcdKey := fmt.Sprintf("%s/versions/%s/%s/%s", baseKey, version, cfg.group, cfg.key)

		revision, err = s.etcd.Put(ctx, etcdKey, cfg.value)
		if err != nil {
			return nil, err
		}
	}

	// 写meta(meta为版本说明)
	meta := model.ConfigMeta{
		Version:   version,
		Revision:  revision,
		Publisher: req.Publisher,
		Comment:   req.Comment,
		CreatedAt: time.Now().Unix(),
	}

	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}

	revision, err = s.etcd.Put(ctx, fmt.Sprintf("%s/meta/%s", baseKey, version), string(metaBytes))
	if err != nil {
		return nil, err
	}

	// 更新current
	revision, err = s.etcd.Put(ctx, baseKey+"/current", version)
	if err != nil {
		return nil, err
	}

	// 更新version_counter
	revision, err = s.etcd.Put(ctx, counterKey, strconv.Itoa(nextNumber))
	if err != nil {
		return nil, err
	}

	// 返回结果
	return &model.PublishConfigResponse{
		App:      req.App,
		Env:      req.Env,
		Version:  version,
		Revision: revision,
	}, nil
}

func splitConfigKey(configKey string) (string, string, error) {
	configKey = strings.TrimSpace(configKey)
	parts := strings.SplitN(configKey, ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid config key: %s", configKey)
	}

	group := strings.TrimSpace(parts[0])
	key := strings.TrimSpace(parts[1])
	if group == "" || key == "" {
		return "", "", fmt.Errorf("invalid config key: %s", configKey)
	}

	return group, key, nil
}

func (s *ConfigService) GetCurrent(ctx context.Context, app, env string) (*model.GetConfigResponse, error) {
	if app == "" {
		return nil, fmt.Errorf("app is required")
	}

	if env == "" {
		return nil, fmt.Errorf("env is required")
	}

	baseKey := fmt.Sprintf("/config/%s/%s", app, env)

	currentKey := baseKey + "/current"
	version, exists, _, err := s.etcd.Get(ctx, currentKey)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("config not found")
	}

	versionPrefix := fmt.Sprintf("%s/versions/%s/", baseKey, version)
	kvs, _, err := s.etcd.GetPrefix(ctx, versionPrefix)
	if err != nil {
		return nil, err
	}
	if len(kvs) == 0 {
		return nil, fmt.Errorf("config not found")
	}

	configs := make(map[string]string, len(kvs))
	for _, kv := range kvs {
		configKey, err := parseConfigKey(string(kv.Key), versionPrefix)
		if err != nil {
			return nil, err
		}
		configs[configKey] = string(kv.Value)
	}

	return &model.GetConfigResponse{
		App:     app,
		Env:     env,
		Version: version,
		Configs: configs,
	}, nil
}

func parseConfigKey(fullKey, prefix string) (string, error) {
	trimmed := strings.TrimPrefix(fullKey, prefix)
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("invalid config key: %s", fullKey)
	}

	return parts[0] + "." + parts[1], nil
}
