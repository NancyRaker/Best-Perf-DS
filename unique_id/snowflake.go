package snowflake

import (
	"sync"
	"time"
)

/*
  0 00000000000000000000000000000000000000000 0000000000 000000000000
                       时间戳  +                   机器id       序列号
*/
const (
	NODE_BIT_SIZE    byte  = 10                           // 机器id偏移
	SEQ_BIT_SIZE     byte  = 12                           // 序列号偏移量
	NODE_ID_MAX      int64 = -1 ^ (-1 << NODE_BIT_SIZE)   // 机器id最大值
	SEQ_MAX          int64 = -1 ^ (-1 << SEQ_BIT_SIZE)    // 序列号最大值
	TIMESTAMP_OFFSET byte  = NODE_BIT_SIZE + SEQ_BIT_SIZE // 时间戳偏移量
	NODE_ID_OFFSET   byte  = SEQ_BIT_SIZE                 //机器ID偏移量
)

// 起始时间戳 (毫秒数显示)
var EPOCH int64 = 1288834974657 // timestamp 2006-03-21:20:50:14 GMT

// ID 结构
type ID int64

type idGenerator struct {
	m                sync.Mutex
	curWorkTimeStamp int64 // 当前获取id的毫秒时间戳
	nodeId           int64 // 机器id
	seqNum           int64 // 当前序列号
}

func GetMillSecond() int64 {
	return time.Now().UnixNano() / 1e6
}

// new IdGenerator
func NewIdGenerator(nodeId int64) *idGenerator {
	if nodeId > NODE_ID_MAX {
		nodeId = NODE_ID_MAX // 保证不溢出
	}
	return &idGenerator{
		nodeId:           nodeId,
		curWorkTimeStamp: GetMillSecond(),
	}
}

// 获取id
func (ig *idGenerator) Next() ID {
	ig.m.Lock()
	defer ig.m.Unlock()

	// 当前时间
	nowTimeStamp := GetMillSecond()

	if ig.curWorkTimeStamp == nowTimeStamp {
		// 序列号加一
		ig.seqNum++
		// 如果这一毫秒内的序列号用完，就循环等待本毫秒结束
		if ig.seqNum > SEQ_MAX {
			for nowTimeStamp <= ig.curWorkTimeStamp {
				nowTimeStamp = GetMillSecond()
			}
		}
	} else {
		// ig.curWorkTimeStamp 这一毫秒内的序列号用不完也算了，直接用最新的这一毫秒的序列号
		ig.seqNum = 0
	}

	ig.curWorkTimeStamp = nowTimeStamp

	return ID((nowTimeStamp-EPOCH)<<TIMESTAMP_OFFSET | (ig.nodeId << NODE_ID_OFFSET) | (ig.seqNum))
}
