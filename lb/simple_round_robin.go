package lb

type simpleRoundRobin struct {
	roundRobinPos int
}

func newSimleRoundRobin() *simpleRoundRobin {
	return &simpleRoundRobin{}
}

// 简单轮询
func (srr *simpleRoundRobin) GetServer(opt ...string) string {
	var res string

	if srr.roundRobinPos >= len(ips) {
		srr.roundRobinPos = 0
	}

	res = ips[srr.roundRobinPos]
	srr.roundRobinPos++

	return res
}
