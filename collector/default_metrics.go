// Copyright (c) 2021, 2025, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
// Portions Copyright (c) 2016 Seth Miller <seth@sethmiller.me>

package collector

import (
	_ "embed"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

//go:embed default_metrics.toml
var defaultMetricsToml string

//go:embed rac_metrics.toml
var racMetricsToml string

//go:embed asm_metrics.toml
var asmMetricsToml string

//go:embed dg_metrics.toml
var dgMetricsToml string

// 创建通用加载函数减少重复代码
func loadMetrics(tomlData string, logger *slog.Logger) ([]Metric, error) {
	var metrics Metrics
	if _, err := toml.Decode(tomlData, &metrics); err != nil {
		return nil, fmt.Errorf("failed to decode metrics: %w", err)
	}
	return metrics.Metric, nil
}

// DefaultMetrics 重构后方法
func (e *Exporter) DefaultMetrics() Metrics {
	var metrics Metrics

	// 优先处理自定义指标文件
	if e.config.DefaultMetricsFile != "" {
		if _, err := toml.DecodeFile(filepath.Clean(e.config.DefaultMetricsFile), &metrics); err != nil {
			e.logger.Error("failed to load custom metrics file",
				"path", e.config.DefaultMetricsFile,
				"error", err)
		} else {
			e.logger.Info("success to load custom metrics file",
				"path", e.config.DefaultMetricsFile)
			return metrics
		}
	}

	// 加载默认基础指标
	if baseMetrics, err := loadMetrics(defaultMetricsToml, e.logger); err != nil {
		panic(fmt.Sprintf("failed to load custom metrics file: %v", err))
	} else {
		metrics.Metric = baseMetrics
	}

	// 定义指标加载配置
	metricConfigs := []struct {
		enableFlag  bool
		metricsToml string
		metricName  string
	}{
		{e.config.IsRAC, racMetricsToml, "RAC"},
		{e.config.IsASM, asmMetricsToml, "ASM"},
		{e.config.IsDG, dgMetricsToml, "DataGuard"},
	}

	// 批量加载特性指标
	for _, cfg := range metricConfigs {
		if cfg.enableFlag {
			if featureMetrics, err := loadMetrics(cfg.metricsToml, e.logger); err != nil {
				e.logger.Error("failed to load feature metrics",
					"type", cfg.metricName,
					"error", err)
			} else {
				metrics.Metric = append(metrics.Metric, featureMetrics...)
				e.logger.Info("success to load feature metrics",
					"type", cfg.metricName,
					"count", len(featureMetrics))
			}
		}
	}

	return metrics
}
