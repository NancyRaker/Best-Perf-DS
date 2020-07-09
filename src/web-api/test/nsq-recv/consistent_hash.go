/*
@Time : 2019/9/10 17:30 
@Author : yanKoo
@File : consistent_hash
@Software: GoLand
@Description:
*/
package main

import (
	"crypto/sha1"

	"math"
	"sort"
	"strconv"
	"sync"
)

const (
	//虚拟节点是实际节点在hash空间的复制品，一个时间节点可对应多个虚拟节点。
	//运用虚拟节点增加在hash空间的节点数量，实现平衡性。尽可能让hash的结果分布到所有缓冲区。使空间尽可能都得到利用。
	VirualSpots = 100
)

//节点信息
type node struct {
	nodeKey   string
	nodeValue uint32
}

type nodesArray []node

func (ns nodesArray) Len() int {
	return len(ns)
}

func (ns nodesArray) Less(i, j int) bool {
	return ns[i].nodeValue < ns[j].nodeValue
}

func (ns nodesArray) Swap(i, j int) {
	ns[i], ns[j] = ns[j], ns[i]
}

func (ns nodesArray) Sort() {
	sort.Sort(ns)
}

//环形hash空间
type RingHash struct {
	vSpots int
	nodes  nodesArray
	weight map[string]int
	mu     sync.RWMutex
}

func NewRingHash(spots int) *RingHash {
	if spots <= 0 {
		spots = VirualSpots
	}
	h := &RingHash{
		vSpots: spots,
		weight: make(map[string]int),
	}
	return h
}

/**
* 增加真实节点集
* @param nodes 节点集  key  权重
 */
func (rh *RingHash) AddNodes(nodes map[string]int) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	for k, w := range nodes {
		rh.weight[k] = w
	}
	rh.rgHashSpace()
}

/**
* 增加真实节点
* @param k key
* @param w 权重
 */
func (rh *RingHash) AddNode(k string, w int) {
	rh.mu.Lock()
	rh.mu.Unlock()
	rh.weight[k] = w
	rh.rgHashSpace()
}

/**
* 移除真实节点
 */
func (rh *RingHash) RemoveNode(k string) {
	rh.mu.Lock()
	rh.mu.Unlock()
	delete(rh.weight, k)
	rh.rgHashSpace()
}

/**
* 更新节点权重
* @param k key
* @param w 权重
 */
func (rh *RingHash) UpdateNode(k string, w int) {
	rh.mu.Lock()
	rh.mu.Unlock()
	rh.weight[k] = w
	rh.rgHashSpace()
}

/**
* 重新生成有序hash空间
 */
func (rh *RingHash) rgHashSpace() {
	var tw int //总权重和
	for _, w := range rh.weight {
		tw += w
	}

	rh.nodes = nodesArray{}
	for k, w1 := range rh.weight {
		//首先计算虚拟节点数量
		num := int(math.Floor(float64(w1) / float64(tw) * float64(rh.vSpots)))
		for i := 0; i < num; i++ {
			vkey := k + ":" + strconv.Itoa(i)
			n := node{
				nodeKey:   k,
				nodeValue: rh.generateHash(vkey),
			}
			rh.nodes = append(rh.nodes, n)
		}
	}
	rh.nodes.Sort()
}

/**
* hash生成
* @param str
 */
func (rh *RingHash) generateHash(str string) uint32 {
	hash := sha1.New()
	hash.Write([]byte(str))
	hashBytes := hash.Sum(nil)
	var bs []byte = hashBytes[6:10]
	if len(bs) < 4 {
		return 0
	}
	v := (uint32(bs[3]) << 24) | (uint32(bs[2]) << 16) | (uint32(bs[1]) << 8) | (uint32(bs[0]))
	hash.Reset()
	return v
}

/**
* 根据key获取分配的节点
* @param k
 */
func (rh *RingHash) GetNodeKey(k string) string {
	rh.mu.RLock()
	defer rh.mu.RUnlock()
	if len(rh.nodes) == 0 {
		return ""
	}

	v := rh.generateHash(k)
	i := sort.Search(len(rh.nodes), func(i int) bool { return rh.nodes[i].nodeValue >= v })

	if i == len(rh.nodes) {
		i = 0
	}
	return rh.nodes[i].nodeKey
}

func (rh *RingHash) NodeLens() int {
	return len(rh.nodes)
}
