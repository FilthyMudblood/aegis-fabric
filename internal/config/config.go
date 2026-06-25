package config

import (
	"os"
	"strconv"
	"strings"
)

type RunMode string

const (
	ModeEnterpriseMesh RunMode = "enterprise-mesh" // 仅启用 AFP-Core (内部拥塞控制)
	ModeOpenExchange   RunMode = "open-exchange"   // 启用 AFP-Core + Network (公网零信任)
)

type CoreConfig struct {
	MaxToolConcurrency int     // 触发 Tool Storm 预警的并发工具调用数
	MemoryWarnRatio    float64 // 上下文爆炸预警线 (例如 0.8)
	OOMPanicRatio      float64 // 物理熔断线 (例如 0.95，防止 K8s OOMKill)
	MaxContextBytes    uint64  // SDK 上报的应用层上下文上限（字节）
	EntropyLimit       float64 // 预防性熔断红线 (AFP_ENTROPY_LIMIT)
	MaxRecursionDepth  uint32  // 递归深度物理红线 (AFP_MAX_RECURSION_DEPTH)
}

type SidecarConfig struct {
	Mode RunMode
	Core CoreConfig
}

// LoadEnvConfig loads the deterministic run mode from Kubernetes ConfigMaps/Env.
func LoadEnvConfig() *SidecarConfig {
	mode := ModeEnterpriseMesh // 默认为企业内网模式
	if strings.ToLower(os.Getenv("AFP_RUN_MODE")) == string(ModeOpenExchange) {
		mode = ModeOpenExchange
	}

	return &SidecarConfig{
		Mode: mode,
		Core: CoreConfig{
			MaxToolConcurrency: 50,
			MemoryWarnRatio:    0.75,
			OOMPanicRatio:      0.90,
			MaxContextBytes:    envUint64OrDefault("AFP_MAX_CONTEXT_BYTES", 512*1024*1024),
			EntropyLimit:       envFloat64OrDefault("AFP_ENTROPY_LIMIT", 0.95),
			MaxRecursionDepth:  uint32(envUint64OrDefault("AFP_MAX_RECURSION_DEPTH", 10)),
		},
	}
}

func envFloat64OrDefault(key string, fallback float64) float64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fallback
	}
	return v
}

func envUint64OrDefault(key string, fallback uint64) uint64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return fallback
	}
	return v
}
