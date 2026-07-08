package service

import (
	"context"

	"go.uber.org/zap"
)

type ConfigDiffEntry struct {
	Key      string `json:"key"`
	OldValue string `json:"old_value,omitempty"`
	NewValue string `json:"new_value,omitempty"`
	Type     string `json:"type"`
}

type ConfigChangeSummary struct {
	Added   []ConfigDiffEntry `json:"added"`
	Removed []ConfigDiffEntry `json:"removed"`
	Changed []ConfigDiffEntry `json:"changed"`
}

type NotifyAgent struct {
	logger *zap.Logger
	audit  *AuditLogger
}

func NewNotifyAgent(logger *zap.Logger, audit *AuditLogger) *NotifyAgent {
	return &NotifyAgent{
		logger: logger,
		audit:  audit,
	}
}

func (a *NotifyAgent) NotifyConfigChange(ctx context.Context, app, env, fromVersion, toVersion string, revision int64, previous, current map[string]string) {
	if a == nil || a.logger == nil {
		return
	}

	summary := BuildConfigChangeSummary(previous, current)

	a.logger.Info("notify agent detected config change",
		zap.String("app", app),
		zap.String("env", env),
		zap.String("from_version", fromVersion),
		zap.String("to_version", toVersion),
		zap.Int64("revision", revision),
		zap.Int("added_count", len(summary.Added)),
		zap.Int("removed_count", len(summary.Removed)),
		zap.Int("changed_count", len(summary.Changed)),
		zap.Any("added", summary.Added),
		zap.Any("removed", summary.Removed),
		zap.Any("changed", summary.Changed),
	)

	a.audit.Log(ctx, "notify_config_change",
		zap.String("app", app),
		zap.String("env", env),
		zap.String("from_version", fromVersion),
		zap.String("to_version", toVersion),
		zap.Int64("revision", revision),
		zap.Int("added_count", len(summary.Added)),
		zap.Int("removed_count", len(summary.Removed)),
		zap.Int("changed_count", len(summary.Changed)),
	)
}

func BuildConfigChangeSummary(previous, current map[string]string) ConfigChangeSummary {
	summary := ConfigChangeSummary{
		Added:   make([]ConfigDiffEntry, 0),
		Removed: make([]ConfigDiffEntry, 0),
		Changed: make([]ConfigDiffEntry, 0),
	}

	for key, oldValue := range previous {
		newValue, ok := current[key]
		if !ok {
			summary.Removed = append(summary.Removed, ConfigDiffEntry{
				Key:      key,
				OldValue: oldValue,
				Type:     "removed",
			})
			continue
		}

		if oldValue != newValue {
			summary.Changed = append(summary.Changed, ConfigDiffEntry{
				Key:      key,
				OldValue: oldValue,
				NewValue: newValue,
				Type:     "changed",
			})
		}
	}

	for key, newValue := range current {
		if _, ok := previous[key]; ok {
			continue
		}

		summary.Added = append(summary.Added, ConfigDiffEntry{
			Key:      key,
			NewValue: newValue,
			Type:     "added",
		})
	}

	return summary
}
