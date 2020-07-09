/*
@Time : 2019/10/25 11:20 
@Author : yanKoo
@File : priority_queue
@Software: GoLand
@Description:
*/
// 此示例演示使用堆接口构建的优先级队列。
package main

import (
"container/heap"
"fmt"
)

// item是我们在优先队列中管理的item。
type Item struct {
	value    string // item的值；任意取值。
	priority int    // 队列中项目的优先级。
	// 该参数是更新所需的，由heap.Interface方法维护。
	index int // 堆中项目的参数。
}

// PriorityQueue实现堆。接口并保存项目。
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// 我们希望Pop给我们最高的优先权，而不是最低的优先权，所以我们使用的比这里更大。
	return pq[i].priority > pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// 更新修改队列中某个项目的优先级和值。
func (pq *PriorityQueue) update(item *Item, value string, priority int) {
	item.value = value
	item.priority = priority
	heap.Fix(pq, item.index)
}

// 本示例使用某些项目创建PriorityQueue，添加和操作项目，
// 然后按优先顺序删除这些项目。
func main() {
	// 一些项目和它们的优先级。
	items := map[string]int{
		"banana": 3, "apple": 2, "pear": 4,
	}

	// 创建一个优先级队列，将其中的项目，和
	// 建立优先队列（堆）不变量。
	pq := make(PriorityQueue, len(items))
	i := 0
	for value, priority := range items {
		pq[i] = &Item{
			value:    value,
			priority: priority,
			index:    i,
		}
		i++
	}
	heap.Init(&pq)

	// 插入一个新项目，然后修改其优先级。
	item := &Item{
		value:    "orange",
		priority: 1,
	}
	heap.Push(&pq, item)
	pq.update(item, item.value, 5)

	// 取出项目；以优先顺序递减排列。
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*Item)
		fmt.Printf("%.2d:%s ", item.priority, item.value)
	}
}