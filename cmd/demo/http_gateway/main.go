package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/FilthyMudblood/aegis-fabric/internal/config"
	"github.com/FilthyMudblood/aegis-fabric/internal/core"
	"github.com/FilthyMudblood/aegis-fabric/internal/dataplane"
	afpv1 "github.com/FilthyMudblood/aegis-fabric/pkg/protocol/v1"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	// 初始化 AFP-Core (企业网格模式)
	cfg := config.LoadEnvConfig()

	// 模拟当前本地系统健康 (内存使用率 10%)
	osProvider := &core.MockMetricsProvider{SimulatedUsage: 0.1}
	entropy := core.NewEntropyMonitor(&cfg.Core, osProvider)

	sea := dataplane.NewSingleExecutionAuthority(cfg, entropy, func() uint64 { return uint64(time.Now().Unix()) }, nil)

	// 构建 HTTP 中间件
	http.HandleFunc("/api/v1/agent/invoke", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("HTTP Gateway received request", "method", r.Method, "path", r.URL.Path)

		// 1. 从标准的 HTTP Header 中提取 AFP 治理元数据
		traceID := r.Header.Get("X-AFP-Trace-ID")
		if traceID == "" {
			traceID = "default-http-trace"
		}

		depthStr := r.Header.Get("X-AFP-Recursion-Depth")
		var depth uint32
		if d, err := strconv.ParseUint(depthStr, 10, 32); err == nil {
			depth = uint32(d)
		}

		cvpStr := r.Header.Get("X-AFP-CVP")
		var cvp float32 = 0.5 // 默认中立信誉
		if c, err := strconv.ParseFloat(cvpStr, 32); err == nil {
			cvp = float32(c)
		}

		// 2. 组装 Protocol Buffers 数据面头
		header := &afpv1.GovernanceHeader{
			PacketId:       uint64(time.Now().UnixNano()),
			TraceId:        traceID,
			RecursionDepth: depth,
			CvpScore:       cvp,
			EntropyLoad:    &afpv1.EntropyLoad{ResourceAsymmetryRatio: 0.1},
		}

		// 3. 将 L7 请求交由底层物理内核 (SingleExecutionAuthority) 裁决
		peerDID := r.Header.Get("X-AFP-Source-DID")
		if peerDID == "" {
			peerDID = "did:afp:http:anonymous"
		}

		err := sea.HandleIngress(r.Context(), peerDID, header)
		if err != nil {
			switch {
			case errors.Is(err, dataplane.ErrStrangerTaxFailed):
				http.Error(w, "stranger tax rejected", http.StatusForbidden)
			case errors.Is(err, dataplane.ErrProbationReject):
				http.Error(w, "probation reject", http.StatusTooManyRequests)
			case errors.Is(err, dataplane.ErrTopologyIsolated):
				http.Error(w, "topology isolated", http.StatusServiceUnavailable)
			default:
				if err.Error() == "afp-core: recursion depth exceeded physical limit, network loop detected" {
					http.Error(w, err.Error(), http.StatusLoopDetected)
				} else if err.Error() == "afp-core: critical context explosion or tool storm detected, circuit breaker open" {
					http.Error(w, err.Error(), http.StatusServiceUnavailable)
				} else {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			}
			return
		}

		// 通过治理后，业务侧在此继续处理（这里返回 mock 成功响应）
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"ok":true,"trace_id":"%s","peer":"%s"}`, traceID, peerDID)))
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	addr := os.Getenv("AFP_HTTP_GATEWAY_ADDR")
	if addr == "" {
		addr = "0.0.0.0:8082"
	}
	slog.Info("HTTP Gateway started", "address", addr)
	if err := http.ListenAndServe(addr, nil); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("HTTP Gateway terminated", "error", err)
		os.Exit(1)
	}
}
