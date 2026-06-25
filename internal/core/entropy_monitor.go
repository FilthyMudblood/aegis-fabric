package core

import (
	"sync/atomic"

	"github.com/FilthyMudblood/aegis-fabric/internal/config"
)

type EntropyMonitor struct {
	cfg                    *config.CoreConfig
	activeToolCalls        int32
	reportedRecursionDepth uint32
	reportedContextBytes   uint64
	osProvider             OSMetricsProvider
}

func NewEntropyMonitor(cfg *config.CoreConfig, provider OSMetricsProvider) *EntropyMonitor {
	if provider == nil {
		provider = &MockMetricsProvider{SimulatedUsage: 0.1} // Default fallback
	}
	return &EntropyMonitor{cfg: cfg, osProvider: provider}
}

func (em *EntropyMonitor) IncrementToolCall() { atomic.AddInt32(&em.activeToolCalls, 1) }
func (em *EntropyMonitor) DecrementToolCall() { atomic.AddInt32(&em.activeToolCalls, -1) }

// ReportApplicationState is the SDK write path into local entropy accounting.
func (em *EntropyMonitor) ReportApplicationState(recursionDepth uint32, contextMemoryBytes uint64) {
	atomic.StoreUint32(&em.reportedRecursionDepth, recursionDepth)
	atomic.StoreUint64(&em.reportedContextBytes, contextMemoryBytes)
}

func (em *EntropyMonitor) ReportedRecursionDepth() uint32 {
	return atomic.LoadUint32(&em.reportedRecursionDepth)
}

func (em *EntropyMonitor) CalculateCoreEntropy() float32 {
	return em.combinedEntropy(1)
}

// CalculatePreFlightEntropy folds SDK intent burst (estimated_tasks) into entropy.
func (em *EntropyMonitor) CalculatePreFlightEntropy(estimatedTasks uint32) float32 {
	if estimatedTasks == 0 {
		estimatedTasks = 1
	}
	return em.combinedEntropy(estimatedTasks)
}

func (em *EntropyMonitor) combinedEntropy(estimatedTasks uint32) float32 {
	tools := float64(atomic.LoadInt32(&em.activeToolCalls))
	toolPressure := tools / float64(em.cfg.MaxToolConcurrency)
	if toolPressure > 1.0 {
		toolPressure = 1.0
	}

	memUsage := em.osProvider.GetMemoryUsageRatio()
	memPressure := memUsage / em.cfg.OOMPanicRatio
	if memPressure > 1.0 {
		memPressure = 1.0
	}

	var contextPressure float64
	if em.cfg.MaxContextBytes > 0 {
		contextPressure = float64(atomic.LoadUint64(&em.reportedContextBytes)) / float64(em.cfg.MaxContextBytes)
		if contextPressure > 1.0 {
			contextPressure = 1.0
		}
	}

	burstPressure := float64(estimatedTasks) / float64(em.cfg.MaxToolConcurrency)
	if burstPressure > 1.0 {
		burstPressure = 1.0
	}

	maxPressure := toolPressure
	if memPressure > maxPressure {
		maxPressure = memPressure
	}
	if contextPressure > maxPressure {
		maxPressure = contextPressure
	}
	if burstPressure > maxPressure {
		maxPressure = burstPressure
	}
	return float32(maxPressure)
}
