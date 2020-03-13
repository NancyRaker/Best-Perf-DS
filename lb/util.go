package lb

import (
	"math/rand"
	"time"
)


// 获取随机数
func getRandom(num int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(num)
}
