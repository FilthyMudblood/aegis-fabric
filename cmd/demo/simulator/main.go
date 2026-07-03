package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"text/tabwriter"
	"time"
)

const (
	TotalEpochs    = 100
	NetworkSize    = 500
	MaliciousRate  = 0.05
	SimulationRuns = 1000 // 蒙特卡洛模拟次数
)

type AgentNode struct {
	ID          int
	IsMalicious bool
	IsDead      bool
	Load        float64
	CVP         float64
}

type SimulationNetwork struct {
	Mode  string
	Nodes []*AgentNode
	rng   *rand.Rand // 隔离的随机数生成器，确保并发安全
}

func NewNetwork(mode string, seed int64) *SimulationNetwork {
	rng := rand.New(rand.NewSource(seed))
	nodes := make([]*AgentNode, NetworkSize)
	for i := 0; i < NetworkSize; i++ {
		nodes[i] = &AgentNode{
			ID:          i,
			IsMalicious: rng.Float64() < MaliciousRate,
			IsDead:      false,
			Load:        0.0,
			CVP:         1.0,
		}
	}
	return &SimulationNetwork{Mode: mode, Nodes: nodes, rng: rng}
}

func (net *SimulationNetwork) Step() (alive int, avgLoad float64) {
	aliveNodes := make([]*AgentNode, 0)
	for _, n := range net.Nodes {
		if !n.IsDead {
			n.Load *= 0.5
			aliveNodes = append(aliveNodes, n)
		}
	}

	if len(aliveNodes) == 0 {
		return 0, 0.0
	}

	for _, src := range aliveNodes {
		for i := 0; i < 3; i++ {
			target := aliveNodes[net.rng.Intn(len(aliveNodes))]
			if src.ID == target.ID { continue }

			requestLoad := 0.1
			if src.IsMalicious {
				requestLoad = 0.8
			}

			if net.Mode == "Baseline" {
				target.Load += requestLoad
			} else if net.Mode == "AFP" {
				// 1. 拓扑隔离防线 (CVP 破产直接拒绝路由)
				if target.CVP < 0.3 {
					continue
				}

				// 2. 预防性熔断防线 (Preemptive Circuit Breaker)
				// 评估：当前负载 + 本次请求带来的物理熵压 是否会击穿红线
				if target.Load+requestLoad > 0.95 {
					// 触发熔断：拒收请求，并对试图打挂自己的源头进行严厉的 CVP 惩罚
					if src.IsMalicious {
						src.CVP -= 0.5 // Asymmetric Hysteresis: 重罚
					} else {
						src.CVP -= 0.1 // 正常节点的偶发拥塞：轻罚
					}
					continue
				}

				// 3. 陌生人税与恶意载荷识别
				// 如果单次请求的熵压过大（明显是外部性爆炸），即使能吃下也要惩罚其信誉
				if requestLoad > 0.5 {
					src.CVP -= 0.3
				}

				// 安全放行
				target.Load += requestLoad
			}
		}
	}

	totalLoad := 0.0
	aliveCount := 0
	for _, n := range net.Nodes {
		if !n.IsDead {
			if n.Load >= 1.0 {
				n.IsDead = true
			} else {
				aliveCount++
				totalLoad += n.Load
			}
		}
	}

	if aliveCount > 0 {
		avgLoad = totalLoad / float64(aliveCount)
	}
	return aliveCount, avgLoad
}

type EpochStat struct {
	AliveSum int
	LoadSum  float64
}

func main() {
	fmt.Printf("Initializing Monte Carlo Engine... Runs: %d, Nodes: %d, Malicious Rate: %.2f\n", SimulationRuns, NetworkSize, MaliciousRate)

	baselineStats := make([]EpochStat, TotalEpochs)
	afpStats := make([]EpochStat, TotalEpochs)
	
	var wg sync.WaitGroup
	var mu sync.Mutex // 保护 Stats 数组

	start := time.Now()

	for run := 0; run < SimulationRuns; run++ {
		wg.Add(1)
		go func(runIdx int) {
			defer wg.Done()
			// 使用纳秒时间戳+runIdx作为强随机种子
			seed := time.Now().UnixNano() + int64(runIdx)
			
			bNet := NewNetwork("Baseline", seed)
			aNet := NewNetwork("AFP", seed)

			var localBStats [TotalEpochs]EpochStat
			var localAStats [TotalEpochs]EpochStat

			for epoch := 0; epoch < TotalEpochs; epoch++ {
				bAlive, bLoad := bNet.Step()
				aAlive, aLoad := aNet.Step()
				
				localBStats[epoch] = EpochStat{AliveSum: bAlive, LoadSum: bLoad}
				localAStats[epoch] = EpochStat{AliveSum: aAlive, LoadSum: aLoad}
			}

			mu.Lock()
			for i := 0; i < TotalEpochs; i++ {
				baselineStats[i].AliveSum += localBStats[i].AliveSum
				baselineStats[i].LoadSum += localBStats[i].LoadSum
				afpStats[i].AliveSum += localAStats[i].AliveSum
				afpStats[i].LoadSum += localAStats[i].LoadSum
			}
			mu.Unlock()
		}(run)
	}

	wg.Wait()
	elapsed := time.Since(start)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "Epoch\t| Base Alive (Avg)\t| Base Load (Avg)\t| AFP Alive (Avg)\t| AFP Load (Avg)\t")
	fmt.Fprintln(w, "-----------------------------------------------------------------------------------------")

	for i := 0; i < TotalEpochs; i++ {
		bAliveAvg := float64(baselineStats[i].AliveSum) / float64(SimulationRuns)
		bLoadAvg := baselineStats[i].LoadSum / float64(SimulationRuns)
		
		aAliveAvg := float64(afpStats[i].AliveSum) / float64(SimulationRuns)
		aLoadAvg := afpStats[i].LoadSum / float64(SimulationRuns)

		fmt.Fprintf(w, "T=%03d\t| %.2f\t| %.4f\t| %.2f\t| %.4f\t\n",
			i+1, bAliveAvg, bLoadAvg, aAliveAvg, aLoadAvg)
	}
	w.Flush()
	fmt.Printf("\nSimulation completed in %v\n", elapsed)
}
