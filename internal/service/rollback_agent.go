package service

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

type RollbackAgent struct {
	logger *zap.Logger
	audit  *AuditLogger
}

func NewRollbackAgent(logger *zap.Logger, audit *AuditLogger) *RollbackAgent {
	return &RollbackAgent{
		logger: logger,
		audit:  audit,
	}
}

func (a *RollbackAgent) Evaluate(ctx context.Context, app, env, version string, configs map[string]string) error {
	if a == nil {
		return nil
	}

	if err := validateAgentConfigRules(env, configs); err != nil {
		if a.logger != nil {
			a.logger.Warn("rollback agent detected risky config",
				zap.String("app", app),
				zap.String("env", env),
				zap.String("version", version),
				zap.Error(err),
			)
		}
		return err
	}

	if a.logger != nil {
		a.logger.Info("rollback agent validation passed",
			zap.String("app", app),
			zap.String("env", env),
			zap.String("version", version),
		)
	}

	return nil
}

func (a *RollbackAgent) RecordRollback(ctx context.Context, app, env, fromVersion, failedVersion, newVersion, reason string) {
	if a == nil {
		return
	}

	if a.logger != nil {
		a.logger.Warn("rollback agent executed auto rollback",
			zap.String("app", app),
			zap.String("env", env),
			zap.String("from_version", fromVersion),
			zap.String("failed_version", failedVersion),
			zap.String("new_version", newVersion),
			zap.String("reason", reason),
		)
	}

	a.audit.Log(ctx, "auto_rollback",
		zap.String("app", app),
		zap.String("env", env),
		zap.String("from_version", fromVersion),
		zap.String("failed_version", failedVersion),
		zap.String("new_version", newVersion),
		zap.String("reason", reason),
	)
}

func validateAgentConfigRules(env string, configs map[string]string) error {
	requiredKeys := []string{
		"database.host",
		"database.port",
		"redis.addr",
	}

	for _, key := range requiredKeys {
		value := strings.TrimSpace(configs[key])
		if value == "" {
			return fmt.Errorf("required config missing or empty: %s", key)
		}
	}

	portValue := strings.TrimSpace(configs["database.port"])
	port, err := strconv.Atoi(portValue)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("database.port must be a valid port: %s", portValue)
	}

	redisAddr := strings.TrimSpace(configs["redis.addr"])
	host, portText, err := net.SplitHostPort(redisAddr)
	if err != nil || strings.TrimSpace(host) == "" {
		return fmt.Errorf("redis.addr must be in host:port format: %s", redisAddr)
	}

	redisPort, err := strconv.Atoi(portText)
	if err != nil || redisPort < 1 || redisPort > 65535 {
		return fmt.Errorf("redis.addr port must be valid: %s", redisAddr)
	}

	if strings.EqualFold(env, "prod") {
		hostValue := strings.TrimSpace(configs["database.host"])
		if hostValue == "127.0.0.1" || strings.EqualFold(hostValue, "localhost") {
			return fmt.Errorf("database.host cannot point to localhost in prod: %s", hostValue)
		}
	}

	return nil
}
