package core

import (
	"sync/atomic"

	"github.com/FilthyMudblood/aegis-fabric/internal/config"
)

type EntropyMonitor struct {
	cfg             *config.CoreConfig
	activeToolCalls int32
	osProvider      OSMetricsProvider
}

func NewEntropyMonitor(cfg *config.CoreConfig, provider OSMetricsProvider) *EntropyMonitor {
	if provider == nil {
		provider = &MockMetricsProvider{SimulatedUsage: 0.1} // Default fallback
	}
	return &EntropyMonitor{cfg: cfg, osProvider: provider}
}

func (em *EntropyMonitor) IncrementToolCall() { atomic.AddInt32(&em.activeToolCalls, 1) }
func (em *EntropyMonitor) DecrementToolCall() { atomic.AddInt32(&em.activeToolCalls, -1) }

func (em *EntropyMonitor) CalculateCoreEntropy() float32 {
	tools := float64(atomic.LoadInt32(&em.activeToolCalls))
	toolPressure := tools / float64(em.cfg.MaxToolConcurrency)
	if toolPressure > 1.0 { toolPressure = 1.0 }

	// 实时读取 OS 接口，而不是依赖外部传值
	memUsage := em.osProvider.GetMemoryUsageRatio()
	memPressure := memUsage / em.cfg.OOMPanicRatio
	if memPressure > 1.0 { memPressure = 1.0 }

	if toolPressure > memPressure { return float32(toolPressure) }
	return float32(memPressure)
}
