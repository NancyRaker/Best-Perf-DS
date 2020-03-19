package hashring

import (
	"crypto/md5"
	"hash"
	"log"
	"math"
	"sort"
	"strconv"
	"sync"
)

const (
	defaultVirtualSpots = 400
)

type node struct {
	key   string
	value uint32
}

type nodesArray []node

func (na nodesArray) Len() int {
	return len(na)
}

func (na nodesArray) Less(i, j int) bool {
	return na[i].value < na[j].value
}

func (na nodesArray) Swap(i, j int) {
	na[i], na[j] = na[j], na[i]
}

func (na *nodesArray) Sort() {
	sort.Sort(na)
}

type hashRing struct {
	virtualSpots int
	nodes        nodesArray
	weights      map[string]int
	hasher       hash.Hash
	mu           sync.RWMutex
}

func NewHashRing(virtualSpots ...int) *hashRing {
	if len(virtualSpots) < 0 || len(virtualSpots) > 1 {
		log.Fatalf("virtualSpots: %+v is invalid", virtualSpots)
		return nil
	}

	h := &hashRing{
		weights: map[string]int{},
		hasher:  md5.New(),
	}
	for _, vs := range virtualSpots {
		h.virtualSpots = vs
	}

	return h
}

func (hr *hashRing) AddNodes(nodeWeight map[string]int) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	for nodeKey, weight := range nodeWeight {
		hr.weights[nodeKey] = weight
	}
	hr.modify()
}

func (hr *hashRing) RemoveNodes(keys []string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	for _, key := range keys {
		delete(hr.weights, key)
	}
	hr.modify()
}

func (hr *hashRing) UpdateNode(key string, weight int) {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	hr.weights[key] = weight
	hr.modify()
}

func (hr *hashRing)GetNode(key string) string {
	if key == "" {
		return ""
	}

	hr.mu.RLock()
	defer hr.mu.RUnlock()

	hashKey := hr.hashVal(hr.hashDigest(key))
	index := sort.Search(hr.nodes.Len(), func(i int) bool {
		return hr.nodes[i].value >= hashKey
	})

	if index == hr.nodes.Len() {
		index = 0
	}

	return hr.nodes[index].key
}

func (hr *hashRing) modify() {
	var totalWeight int
	for _, w := range hr.weights {
		totalWeight += w
	}

	// step1. figure out totalWeight spot
	totalVirtualSpots := hr.virtualSpots * len(hr.weights)

	// step2. generate hash ring
	hr.nodes = nodesArray{}
	for key, weight := range hr.weights {
		// steps2.1 generate totalWeight nodes of single physical node
		spots := int(math.Floor(float64(weight) / float64(totalWeight) * float64(totalVirtualSpots)))
		for i := 0; i < spots; i++ {
			nodeKey := key + "-" + strconv.Itoa(i)
			// generate hash digest
			hKey := hr.hashDigest(nodeKey)

			node := node{
				key:key,
				value:hr.hashVal(hKey),
			}

			hr.nodes = append(hr.nodes, node)
		}
	}
	// step3. sort node
	hr.nodes.Sort()
}

func (hr *hashRing) hashVal(bKey []byte) uint32 {
	if len(bKey) < 4 {
		log.Fatalf("hash key %+v is invalid", bKey)
		return 0
	}
	return (uint32(bKey[3]) << 24) | (uint32(bKey[2]) << 16) | (uint32(bKey[1]) << 8) | (uint32(bKey[0]))
}


func (hr *hashRing) hashDigest(key string) []byte {
	defer hr.hasher.Reset()
	return hr.hasher.Sum([]byte(key))
}
