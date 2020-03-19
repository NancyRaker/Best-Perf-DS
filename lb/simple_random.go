package lb

type simpleRandom struct{}

func newSimpleRandom() *simpleRandom {
	return &simpleRandom{}
}

// 简单随机
func (sr *simpleRandom) GetServer(opt ...string) string {
	return ips[getRandom(len(ips))]
}