package core

import (
	"runtime"
)

// OSMetricsProvider defines the contract for reading physical OS limits.
type OSMetricsProvider interface {
	GetMemoryUsageRatio() float64
}

// MockMetricsProvider is used for local development and simulation.
type MockMetricsProvider struct {
	SimulatedUsage float64
}

func (m *MockMetricsProvider) GetMemoryUsageRatio() float64 {
	return m.SimulatedUsage
}

// CgroupsMetricsProvider will be used in actual Kubernetes deployments.
type CgroupsMetricsProvider struct{}

func (c *CgroupsMetricsProvider) GetMemoryUsageRatio() float64 {
	// 真实部署时将读取 /sys/fs/cgroup/memory.current / memory.max
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// Fallback to Go runtime stats for now
	return float64(m.Alloc) / float64(1024*1024*512) // 假设 512MB 为容器限额
}
