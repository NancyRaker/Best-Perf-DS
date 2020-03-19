package lb

import (
	"testing"
)

func TestLoadBalance_GetServer(t *testing.T) {
	loadBalance := NewLoadBalance(
		[]string{
			"192.168.10.1:2202",
			"192.168.10.2:2202",
			"192.168.10.3:2202",},
		[]int{
			5,
			10,
			1,})

	var resCount = []int{0,0,0}
	for i := 0; i < 128; i++ {
		ip := loadBalance.GetServer(WEIGHT_ROUND_ROBIN, "")
		if ip == "192.168.10.1:2202" {
			resCount[0]++
		}else if ip == "192.168.10.2:2202" {
			resCount[1]++
		}else if ip == "192.168.10.3:2202" {
			resCount[2]++
		}
	}

	t.Logf("weight round robin res: %+v", resCount)

	t.Logf("source hash res: %+v", loadBalance.GetServer(SOURCE_HASH, "127.0.0.1"))
	t.Logf("source hash res: %+v", loadBalance.GetServer(SOURCE_HASH, "127.0.0.1"))
	t.Logf("source hash res: %+v", loadBalance.GetServer(SOURCE_HASH, "127.0.0.1"))
}