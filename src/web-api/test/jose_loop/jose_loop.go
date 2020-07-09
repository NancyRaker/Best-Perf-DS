/*
@Time : 2019/11/4 15:05 
@Author : yanKoo
@File : jose_loop
@Software: GoLand
@Description:
*/
package jose_loop

import "fmt"

type joseRing struct {
	head   *node
	target int
	tail   *node
}

type node struct {
	value int
	next  *node
}

/*
element     1 2 3 4 5 6 7
target    2
killCount   0 1
temp          t
*/
func NewJoseLoop(n, target int) *joseRing {
	var head = &node{value: 1}
	var dummy = head
	for i := 2; i <= n; i ++ {
		head.next = &node{value: i}
		head = head.next
	}
	head.next = dummy
	return &joseRing{target: target, head: dummy, tail: head}
}

func (jr *joseRing) printElement() {
	var dummy = jr.head
	var temp = jr.head

	fmt.Printf(" %d", temp.value)
	temp = temp.next
	for dummy != temp {
		fmt.Printf(" %d", temp.value)
		temp = temp.next
	}
	fmt.Println()
}

func (jr joseRing) StartKill() int {
	if jr.target == 1 {
		return jr.tail.value
	}

	var node = jr.head
	var killCount int
	for node.next != node {
		if killCount == jr.target-2 {
			// 移除当前的下一个节点
			next := node.next
			node.next = next.next
			next.next = nil // gc
			// 从下一个node从新计数
			node = node.next
			killCount = 0
			continue
		}
		node = node.next
		if jr.target != 2 {
			killCount++
		}
	}
	return node.value
}
