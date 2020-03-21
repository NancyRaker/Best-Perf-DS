/*
@Time : 2019/11/13 13:53
@Author : yanKoo
@File : bst
@Software: GoLand
@Description: bst
*/
package binary_search_tree

import (
	"math"
	"sync"
)

type treeNode struct {
	element     int
	left, right *treeNode
}

func newTreeNode(e int) *treeNode {
	return &treeNode{
		element: e,
	}
}

type binarySearchTree struct {
	root *treeNode
	size int
	m    sync.RWMutex
}

func NewBST() *binarySearchTree {
	return &binarySearchTree{}
}

func (bst *binarySearchTree) Size() int {
	bst.m.RLock()
	defer bst.m.RUnlock()
	return bst.size
}

func (bst *binarySearchTree) Add(e int) {
	bst.m.Lock()
	defer bst.m.Unlock()
	bst.root = bst.add(bst.root, e)
}

func (bst *binarySearchTree) add(node *treeNode, e int) *treeNode {
	if node == nil {
		bst.size++
		return newTreeNode(e)
	}

	if e < node.element {
		node.left = bst.add(node.left, e)
	} else if e > node.element {
		node.right = bst.add(node.right, e)
	}

	return node
}

func (bst *binarySearchTree) Remove(e int) {
	bst.m.Lock()
	defer bst.m.Unlock()
	bst.root = bst.remove(bst.root, e)
}

func (bst *binarySearchTree) remove(node *treeNode, e int) *treeNode {
	if node == nil {
		return nil
	}

	if e < node.element {
		node.left = bst.remove(node.left, e)
		return node
	} else if e > node.element {
		node.right = bst.remove(node.right, e)
		return node
	} else { // e == node.element
		// 1. 左子树为空
		if node.left == nil {
			rightNode := node.right
			node.right = nil
			bst.size--
			return rightNode
		}

		// 2. 右子树为空
		if node.right == nil {
			leftNode := node.left
			node.left = nil
			bst.size--
			return leftNode
		}

		// 3. 左右子树都不为空
		// 找到待删除的节点的后继（比待删除节点的最小节点），然后用这个后继代替待删除的节点. Hibbard deletion
		successor := bst.minimum(node.right)
		successor.right = bst.removeMinNode(node.right)
		successor.left = node.left
		return successor
	}
}

func (bst *binarySearchTree) RemoveMin() int {
	if bst.Size() == 0 {
		return math.MaxInt32
	}
	bst.m.Lock()
	defer bst.m.Unlock()
	return bst.removeMin(bst.root)
}

// 以root为根的bst的最小值
func (bst *binarySearchTree) minimum(root *treeNode) *treeNode {
	if root.left == nil {
		return root
	}
	return bst.minimum(root.left)
}

// 删除以root为根的bst的最小值，并返回这个值
func (bst *binarySearchTree) removeMin(root *treeNode) int {
	// 获取bst最小元素
	ret := bst.minimum(root).element

	// 删除最小节点
	root = bst.removeMinNode(root)

	// 返回被删除的元素
	return ret
}

// 删除以root为根的bst的最小值，并返回这个位置应该变成的节点
func (bst *binarySearchTree) removeMinNode(node *treeNode) *treeNode {
	if node.left == nil {
		rightNode := node.right
		node.right = nil
		bst.size--
		return rightNode
	}

	node.left = bst.removeMinNode(node.left)
	return node
}

func (bst *binarySearchTree) RemoveMax() int {
	if bst.Size() == 0 {
		return math.MinInt32
	}

	bst.m.Lock()
	defer bst.m.Unlock()
	return bst.removeMax(bst.root)
}

// 以root为根的bst的最小值
func (bst *binarySearchTree) maximum(root *treeNode) *treeNode {
	if root.right == nil {
		return root
	}

	return bst.maximum(root.right)
}

// 删除以root为根的bst中的最大值
func (bst *binarySearchTree) removeMax(root *treeNode) int {
	ret := bst.maximum(root).element
	root = bst.removeMaxNode(root)
	return ret
}

// 删除以root为根的bst的最大值，并返回这个位置应该变成的节点
func (bst *binarySearchTree) removeMaxNode(node *treeNode) *treeNode {
	if node.right == nil {
		leftNode := node.left
		node.left = nil
		bst.size--
		return leftNode
	}

	node.right = bst.removeMaxNode(node.right)
	return node
}

func (bst *binarySearchTree) Search(e int) bool {
	bst.m.RLock()
	defer bst.m.RUnlock()
	return bst.search(bst.root, e)
}

func (bst *binarySearchTree) search(node *treeNode, e int) bool {
	if node == nil {
		return false
	}

	if node.element == e {
		return true
	} else if e < node.element {
		return bst.search(node.left, e)
	} else {
		return bst.search(node.right, e)
	}
}
