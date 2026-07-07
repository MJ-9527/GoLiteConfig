package service

import (
	"GoLiteConfig/internal/etcd"
	"GoLiteConfig/internal/model"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ConfigService struct {
	etcd     *etcd.Client
	watchMgr *WatchManager
}

func NewConfigService(etcdClient *etcd.Client, watchMgr *WatchManager) *ConfigService {
	return &ConfigService{
		etcd:     etcdClient,
		watchMgr: watchMgr,
	}
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

	if s.watchMgr != nil {
		s.watchMgr.Notify(req.App, req.Env)
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
	kvs, revision, err := s.etcd.GetPrefix(ctx, versionPrefix)
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
		App:      app,
		Env:      env,
		Version:  version,
		Revision: revision,
		Configs:  configs,
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

func (s *ConfigService) ListVersions(ctx context.Context, app, env string) (*model.ListVersionsResponse, error) {
	if app == "" {
		return nil, fmt.Errorf("app is required")
	}

	if env == "" {
		return nil, fmt.Errorf("env is required")
	}

	baseKey := fmt.Sprintf("/config/%s/%s", app, env)

	current, exists, _, err := s.etcd.Get(ctx, baseKey+"/current")
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("config not found")
	}

	metaPrefix := baseKey + "/meta/"
	kvs, _, err := s.etcd.GetPrefix(ctx, metaPrefix)
	if err != nil {
		return nil, err
	}
	if len(kvs) == 0 {
		return nil, fmt.Errorf("config not found")
	}

	versions := make([]model.ConfigMeta, 0, len(kvs))
	for _, kv := range kvs {
		var meta model.ConfigMeta
		if err = json.Unmarshal(kv.Value, &meta); err != nil {
			return nil, err
		}
		versions = append(versions, meta)
	}

	sort.Slice(versions, func(i, j int) bool {
		return versionNumber(versions[i].Version) < versionNumber(versions[j].Version)
	})

	return &model.ListVersionsResponse{
		App:      app,
		Env:      env,
		Current:  current,
		Versions: versions,
	}, nil
}

func versionNumber(version string) int {
	version = strings.TrimPrefix(version, "v")
	n, err := strconv.Atoi(version)
	if err != nil {
		return 0
	}
	return n
}

func (s *ConfigService) getConfigByVersion(ctx context.Context, app, env, version string) (map[string]string, error) {
	if app == "" || env == "" || version == "" {
		return nil, fmt.Errorf("invalid rollback target")
	}

	baseKey := fmt.Sprintf("/config/%s/%s", app, env)
	versionPrefix := fmt.Sprintf("%s/versions/%s/", baseKey, version)

	kvs, _, err := s.etcd.GetPrefix(ctx, versionPrefix)
	if err != nil {
		return nil, err
	}
	if len(kvs) == 0 {
		return nil, fmt.Errorf("target version not found")
	}

	configs := make(map[string]string, len(kvs))
	for _, kv := range kvs {
		configKey, err := parseConfigKey(string(kv.Key), versionPrefix)
		if err != nil {
			return nil, err
		}
		configs[configKey] = string(kv.Value)
	}

	return configs, nil
}

func (s *ConfigService) Rollback(ctx context.Context, req model.RollbackRequest) (*model.RollbackResponse, error) {
	if req.App == "" {
		return nil, fmt.Errorf("app is required")
	}

	if req.Env == "" {
		return nil, fmt.Errorf("env is required")
	}

	if req.TargetVersion == "" {
		return nil, fmt.Errorf("target version is required")
	}

	baseKey := fmt.Sprintf("/config/%s/%s", req.App, req.Env)

	fromVersion, exists, _, err := s.etcd.Get(ctx, baseKey+"/current")
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("config not found")
	}

	targetConfig, err := s.getConfigByVersion(ctx, req.App, req.Env, req.TargetVersion)
	if err != nil {
		return nil, err
	}

	comment := req.Comment
	if strings.TrimSpace(comment) == "" {
		comment = fmt.Sprintf("rollback to %s", req.TargetVersion)
	}

	publishResp, err := s.Publish(ctx, model.PublishConfigRequest{
		App:       req.App,
		Env:       req.Env,
		Configs:   targetConfig,
		Publisher: req.Publisher,
		Comment:   comment,
	})
	if err != nil {
		return nil, err
	}

	return &model.RollbackResponse{
		App:           req.App,
		Env:           req.Env,
		FromVersion:   fromVersion,
		TargetVersion: req.TargetVersion,
		NewVersion:    publishResp.Version,
		Revision:      publishResp.Revision,
	}, nil
}

func (s *ConfigService) Watch(ctx context.Context, app, env string, lastRevision int64) (*model.WatchConfigResponse, bool, error) {
	if app == "" {
		return nil, false, fmt.Errorf("app is required")
	}

	if env == "" {
		return nil, false, fmt.Errorf("env is required")
	}

	current, err := s.GetCurrent(ctx, app, env)
	if err != nil {
		return nil, false, err
	}

	if current.Revision > lastRevision {
		return &model.WatchConfigResponse{
			App:      current.App,
			Env:      current.Env,
			Version:  current.Version,
			Revision: current.Revision,
			Configs:  current.Configs,
		}, true, nil
	}

	ch := s.watchMgr.Add(app, env)
	defer s.watchMgr.Remove(app, env, ch)

	timer := time.NewTimer(60 * time.Second)
	defer timer.Stop()

	select {
	case <-ch:
		latest, err := s.GetCurrent(ctx, app, env)
		if err != nil {
			return nil, false, err
		}
		return &model.WatchConfigResponse{
			App:      latest.App,
			Env:      latest.Env,
			Version:  latest.Version,
			Revision: latest.Revision,
			Configs:  latest.Configs,
		}, true, nil

	case <-timer.C:
		return nil, false, nil

	case <-ctx.Done():
		return nil, false, ctx.Err()
	}
}

func (s *ConfigService) DeleteVersions(ctx context.Context, req model.DeleteVersionsRequest) (*model.DeleteVersionsResponse, error) {
	if req.App == "" {
		return nil, fmt.Errorf("app is required")
	}
	if req.Env == "" {
		return nil, fmt.Errorf("env is required")
	}

	versions := normalizeVersions(req.Version, req.Versions)
	if len(versions) == 0 {
		return nil, fmt.Errorf("version or versions is required")
	}

	baseKey := fmt.Sprintf("/config/%s/%s", req.App, req.Env)
	current, exists, _, err := s.etcd.Get(ctx, baseKey+"/current")
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("config not found")
	}

	deleted := make([]string, 0, len(versions))
	skipped := make([]string, 0, len(versions))
	var revision int64

	for _, version := range versions {
		if version == current {
			skipped = append(skipped, version)
			continue
		}

		configPrefix := fmt.Sprintf("%s/versions/%s/", baseKey, version)
		metaKey := fmt.Sprintf("%s/meta/%s", baseKey, version)

		kvs, _, err := s.etcd.GetPrefix(ctx, configPrefix)
		if err != nil {
			return nil, err
		}
		if len(kvs) == 0 {
			return nil, fmt.Errorf("target version not found")
		}

		revision, err = s.etcd.DeletePrefix(ctx, configPrefix)
		if err != nil {
			return nil, err
		}
		revision, err = s.etcd.Delete(ctx, metaKey)
		if err != nil {
			return nil, err
		}

		deleted = append(deleted, version)
	}

	if len(deleted) == 0 && len(skipped) > 0 {
		return nil, fmt.Errorf("cannot delete current version")
	}

	return &model.DeleteVersionsResponse{
		App:      req.App,
		Env:      req.Env,
		Current:  current,
		Deleted:  deleted,
		Skipped:  skipped,
		Revision: revision,
	}, nil
}

func normalizeVersions(single string, multiple []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(multiple)+1)

	add := func(version string) {
		version = strings.TrimSpace(version)
		if version == "" {
			return
		}
		if _, ok := seen[version]; ok {
			return
		}
		seen[version] = struct{}{}
		result = append(result, version)
	}

	add(single)
	for _, version := range multiple {
		add(version)
	}

	return result
}

func (s *ConfigService) ResetConfigs(ctx context.Context) error {
	_, err := s.etcd.DeletePrefix(ctx, "/config/")
	return err
}
