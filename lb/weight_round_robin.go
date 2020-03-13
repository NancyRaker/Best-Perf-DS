package lb

import (
	"errors"
)

type weightRoundRobin struct {
	maxPos     int // 当前取出ip的权重
	total      int // 权重和
	ips        []string
	weights    []int
	curWeights []int
}

func newWeightRoundRobin(ips []string, weights []int) (*weightRoundRobin, error) {
	if ips == nil || weights == nil || len(ips) != len(weights) {
		return nil, errors.New("invalid param")
	}

	var wrr = &weightRoundRobin{
		ips:     ips,
		weights: weights,
		curWeights: append([]int{}, weights...),
	}

	for _, weight :=  range weights {
		wrr.total += weight
	}
	return wrr, nil
}

// 平滑加权轮询 实现LoadBalance接口
func (wrr *weightRoundRobin) GetServer() string {
	// 返回 maxPos上的ip
	var res = wrr.ips[wrr.maxPos]

	// wrr.maxPos 位置的ip权重减去总权重
	wrr.curWeights[wrr.maxPos] -= wrr.total

	// 恢复curWeights
	for i := range wrr.curWeights {
		wrr.curWeights[i] += wrr.weights[i]
	}

	// 更新maxPos
	wrr.updateMaxWeightPos()

	return res
}

// 更新最大ip
func (wrr *weightRoundRobin) updateMaxWeightPos() {
	// 更新最大权重
	var maxWeightPos = wrr.curWeights[0]
	wrr.maxPos = 0
	for i, weight := range wrr.curWeights {
		if weight > maxWeightPos {
			wrr.maxPos = i
			maxWeightPos = weight
		}
	}
}
