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
	baseKey := fmt.Sprintf("/configs/%s/%s", req.App, req.Env)

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
	var revision int64

	for configKey, configValue := range req.Configs {
		group, key, err := splitConfigKey(configKey)
		if err != nil {
			return nil, err
		}

		etcdKey := fmt.Sprintf("%s/versions/%s/%s/%s", baseKey, version, group, key)

		revision, err = s.etcd.Put(ctx, etcdKey, configValue)
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
	parts := strings.SplitN(configKey, ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid config key: %s", configKey)
	}

	return parts[0], parts[1], nil
}
