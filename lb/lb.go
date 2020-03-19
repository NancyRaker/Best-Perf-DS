package lb

type LoadBalanceStrategy interface {
	GetServer(... string) string
}

type loadBalance struct {
	simpleRandom     LoadBalanceStrategy
	weightRandom     LoadBalanceStrategy
	simpleRoundRobin LoadBalanceStrategy
	weightRoundRobin LoadBalanceStrategy
	sourceHash       LoadBalanceStrategy
}

const (
	SIMPLE_RANDOM      = 1
	WEIGHT_RANDOM      = 2
	SIMPLE_ROUND_ROBIN = 3
	WEIGHT_ROUND_ROBIN = 4
	SOURCE_HASH        = 5
)

var ips = []string{
	"192.168.0.1",
	"192.168.0.2",
	"192.168.0.3",
	"192.168.0.4",
	"192.168.0.5",
	"192.168.0.6",
	"192.168.0.7",
	"192.168.0.8",
	"192.168.0.9",
	"192.168.0.10",
}

var weights = []int{100, 2, 3, 4, 5, 6, 7, 8, 9, 10}

var ipsMaps = map[string]int{
	"192.168.0.1":  20,
	"192.168.0.2":  10,
	"192.168.0.3":  5,
	"192.168.0.4":  7,
	"192.168.0.5":  45,
	"192.168.0.6":  3,
	"192.168.0.7":  100,
	"192.168.0.8":  2,
	"192.168.0.9":  3,
	"192.168.0.10": 5,
}

func NewLoadBalance(ips []string, weights []int) *loadBalance {
	wrr, _ := newWeightRoundRobin(ips, weights)
	return &loadBalance{
		simpleRandom:     newSimpleRandom(),
		weightRandom:     newWeightRandom(),
		simpleRoundRobin: newSimleRoundRobin(),
		weightRoundRobin: wrr,
		sourceHash:       newSourceHash(ips),
	}
}

func (lb *loadBalance) GetServer(strategy int, opt string) string {
	var res string
	switch strategy {
	case SIMPLE_RANDOM:
		res = lb.simpleRandom.GetServer()
	case WEIGHT_RANDOM:
		res = lb.weightRandom.GetServer()
	case SIMPLE_ROUND_ROBIN:
		res = lb.simpleRoundRobin.GetServer()
	case WEIGHT_ROUND_ROBIN:
		res = lb.weightRoundRobin.GetServer()
	case SOURCE_HASH:
		res = lb.sourceHash.GetServer(opt)
	}
	return res
}
