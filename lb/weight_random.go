package lb

// 加权随机
type weightRandom struct{}

func newWeightRandom() *weightRandom {
	return &weightRandom{}
}

func (wr *weightRandom) GetServer() string {
	// 1. calc total weight
	var totalWeight int
	for _, value := range ipsMaps {
		totalWeight += value
	}

	// 2. random get offset
	offset := getRandom(totalWeight)
	for ip, weight := range ipsMaps {
		if offset < weight {
			return ip
		}
		offset -= weight
	}

	// 3. 如果权重都一样就是简单随机
	sr := simpleRandom{}
	return sr.GetServer()
}
