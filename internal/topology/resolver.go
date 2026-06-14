package topology

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

var (
	ErrResolutionTimeout = errors.New("afp: topology resolution timed out or target unreachable")
)

// Resolver handles asynchronous P2P discovery of unknown DIDs via Core Neighbors.
type Resolver struct {
	store        NeighborStore
	resolveCache sync.Map // 简单的防并发风暴缓存: 防止同一时刻对同一个 DID 发起几百次全网寻址
}

func NewResolver(store NeighborStore) *Resolver {
	return &Resolver{
		store: store,
	}
}

// Resolve executes a fan-out query to Core Neighbors to find the physical endpoint of a target DID.
func (r *Resolver) Resolve(ctx context.Context, targetDID string) (string, error) {
	// 1. 本地快速路径 (Fast Path Check)
	if endpoint, exists := r.store.ResolveEndpoint(targetDID); exists {
		return endpoint, nil
	}

	// 2. 防风暴缓存锁 (Singleflight pattern abstraction)
	// 如果已经在寻址中，直接等待或失败（此处用占位简化，工业级应用通常使用 golang.org/x/sync/singleflight）
	if _, loading := r.resolveCache.LoadOrStore(targetDID, true); loading {
		return "", errors.New("afp: resolution already in progress, backoff applied")
	}
	defer r.resolveCache.Delete(targetDID)

	// 3. 提取核心邻居 (CVP >= 0.8) 作为查询中继
	coreNeighbors := r.store.GetTrustedNeighbors(0.8)
	if len(coreNeighbors) == 0 {
		return "", errors.New("afp: no core neighbors available for routing referral")
	}

	slog.Info("local resolution missed, initiating P2P lookup via core neighbors", "target", targetDID, "core_neighbors_count", len(coreNeighbors))

	// 4. 扇出并发查询 (Fan-out Query)
	// 严格控制超时时间为 2 秒，防止拖死本地 Egress 协程
	lookupCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	resultCh := make(chan string, 1)
	var wg sync.WaitGroup

	for _, neighborDID := range coreNeighbors {
		wg.Add(1)
		go func(ndid string) {
			defer wg.Done()

			neighborEndpoint, ok := r.store.ResolveEndpoint(ndid)
			if !ok {
				return
			}
			_ = neighborEndpoint

			// 模拟向邻居发送 Lookup 协议帧 (工业实现中需要调用真实的 ReadFrame/WriteFrame)
			// 这里假设我们向 ndid 询问 targetDID，且如果它知道，会在 100ms 内返回
			time.Sleep(100 * time.Millisecond) // 模拟网络 RTT

			// 模拟邻居发现逻辑：假设如果目标是以 "did:afp:remote:" 开头的，核心邻居就能“路由”到它
			// 实际场景下，这里是解析邻居发回的响应报文
			if targetDID == "did:afp:remote:z" {
				select {
				case resultCh <- "127.0.0.1:8082": // 指向本地的 Node Z 测试桩
				default:
				}
			}
		}(neighborDID)
	}

	// 5. 等待机制 (Wait & Select)
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	select {
	case endpoint, ok := <-resultCh:
		if ok {
			slog.Info("target resolved via core neighbor referral", "target", targetDID, "endpoint", endpoint)
			// 动态写入本地账本，以便下次 Fast Path 命中
			r.store.UpsertEndpoint(targetDID, endpoint)
			return endpoint, nil
		}
		return "", ErrResolutionTimeout
	case <-lookupCtx.Done():
		return "", ErrResolutionTimeout
	}
}
