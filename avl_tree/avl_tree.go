/*
@Time : 2019/11/13 19:53
@Author : yanKoo
@File : avl_tree
@Software: GoLand
@Description: avl tree
*/
package avl_tree

import (
	"sync"
)

type treeNode struct {
	key         int
	value       interface{}
	left, right *treeNode
	height      int
}

func newTreeNode(key int, value interface{}) *treeNode {
	return &treeNode{
		key:    key,
		value:  value,
		height: 1, // 一个起始节点的高度为1
	}
}

// 获得节点的高度
func (t *treeNode) getHeight() int {
	if t == nil {
		return 0 // 空树高度为0
	}
	return t.height
}

// 获得节点的平衡因子
func (t *treeNode) getBalanceFactor() int {
	if t == nil {
		return 0
	}
	return t.left.getHeight() - t.right.getHeight()
}

type avlTree struct {
	root *treeNode
	size int
	m    sync.RWMutex
}

func NewAVLTree() *avlTree {
	return &avlTree{}
}

func (avl *avlTree) Size() int {
	avl.m.RLock()
	defer avl.m.RUnlock()
	return avl.size
}

func (avl *avlTree) Add(key int, value interface{}) {
	avl.m.Lock()
	defer avl.m.Unlock()
	avl.root = avl.add(avl.root, key, value)
}

// 对节点node进行向右旋转操作，返回旋转后新的根节点sub
//        node                           sub
//       / \                            /   \
//      sub   T3     向右旋转 (y)       z     node
//     / \       - - - - - - - ->    / \     /   \
//    z   subRight                 T1  T2 subRight T3
//   / \
// T1   T2
func (avl *avlTree) rightRotate(node *treeNode) *treeNode {
	leftSub := node.left
	subRight := leftSub.right

	// right rotate
	leftSub.right = node
	node.left = subRight

	node.height = max(node.left.getHeight(), node.right.getHeight()) + 1
	leftSub.height = max(leftSub.left.getHeight(), leftSub.right.getHeight()) + 1

	return leftSub
}

// 对节点node进行向左旋转操作，返回旋转后新的根节点sub
//    node                                sub
//  /    \                             /      \
// T1    sub      向左旋转 (y)       node       z
//       /  \   - - - - - - - ->   / \        / \
//  subLeft  z                   T1 subLeft T3  T4
//          / \
//         T3 T4
func (avl *avlTree) leftRotate(node *treeNode) *treeNode {
	sub := node.right
	subLeft := sub.left

	sub.left = node
	node.right = subLeft

	node.height = max(node.left.getHeight(), node.right.getHeight()) + 1
	sub.height = max(sub.left.getHeight(), sub.right.getHeight()) + 1

	return sub
}

func (avl *avlTree) add(node *treeNode, key int, value interface{}) *treeNode {
	if node == nil {
		avl.size++
		return newTreeNode(key, value)
	}

	if key < node.key {
		node.left = avl.add(node.left, key, value)
	} else if key > node.key {
		node.right = avl.add(node.right, key, value)
	} else {
		node.value = value
	}

	// 更新height
	node.height = max(node.left.getHeight(), node.right.getHeight()) + 1

	return avl.keepBalance(node)
}

// 调整节点保持avl的平衡性
func (avl *avlTree) keepBalance(node *treeNode) *treeNode {
	// 计算平衡因子
	balanceFactor := node.getBalanceFactor()

	// 维护平衡
	// LL
	if balanceFactor > 1 && node.left.getBalanceFactor() >= 0 {
		// 需要右旋
		return avl.rightRotate(node)
	}

	// RR
	if balanceFactor < -1 && node.right.getBalanceFactor() <= 0 {
		// 需要左旋
		return avl.leftRotate(node)
	}

	// LR
	if abs(balanceFactor) > 1 && node.left.getBalanceFactor() < 0 {
		// 先对左孩子左旋变成LL
		node.left = avl.leftRotate(node.left)
		// 然后在对node右旋
		return avl.rightRotate(node)
	}

	// RL
	if balanceFactor < -1 && node.right.getBalanceFactor() > 0 {
		// 先对右孩子右旋变成RR
		node.right = avl.rightRotate(node.right)
		// 在对node左旋
		return avl.leftRotate(node)
	}

	return node
}

func (avl *avlTree) Remove(e int) {
	avl.m.Lock()
	defer avl.m.Unlock()
	avl.root = avl.remove(avl.root, e)
}

func (avl *avlTree) remove(node *treeNode, key int) *treeNode {
	if node == nil {
		return nil
	}

	var retNode *treeNode
	if key < node.key {
		node.left = avl.remove(node.left, key)
		retNode = node

	} else if key > node.key {
		node.right = avl.remove(node.right, key)
		retNode = node

	} else { // key == node.element
		if node.left == nil { // 1. 左子树为空
			rightNode := node.right
			node.right = nil
			avl.size--
			retNode = rightNode
		} else if node.right == nil { // 2. 右子树为空
			leftNode := node.left
			node.left = nil
			avl.size--
			retNode = leftNode
		} else {
			// 3. 左右子树都不为空
			// 找到待删除的节点的后继（比待删除节点的最小节点），然后用这个后继代替待删除的节点. Hibbard deletion
			successor := avl.minimum(node.right)
			successor.right = avl.remove(node.right, successor.key)
			successor.left = node.left
			retNode = successor
		}
	}

	if retNode == nil {
		return nil
	}

	// 调整节点保持平衡
	return avl.keepBalance(retNode)
}

func (avl *avlTree) IsBST() bool {
	var arr = make([]int, 0)
	inOrder(avl.root, &arr)
	for i := 1; i < len(arr); i++ {
		if arr[i] < arr[i-1] {
			return false
		}
	}
	return true
}

func inOrder(node *treeNode, res *[]int) {
	if node == nil {
		return
	}
	inOrder(node.left, res)
	*res = append(*res, node.key)
	inOrder(node.right, res)
}

// 判断这个avl树是不是真的平衡
func (avl *avlTree) IsBalance() bool {
	return avl.isBalance(avl.root)
}

func (avl *avlTree) isBalance(node *treeNode) bool {
	if node == nil {
		return true
	}

	if abs(node.getBalanceFactor()) > 1 {
		return false
	}

	return avl.isBalance(node.left) && avl.isBalance(node.right)
}

// 以root为根的avl的最小值
func (avl *avlTree) minimum(root *treeNode) *treeNode {
	if root.left == nil {
		return root
	}
	return avl.minimum(root.left)
}

func (avl *avlTree) Get(key int) *treeNode {
	avl.m.RLock()
	defer avl.m.RUnlock()
	return avl.getNode(avl.root, key)
}

func (avl *avlTree) Contains(key int) bool {
	avl.m.RLock()
	defer avl.m.RUnlock()
	return avl.getNode(avl.root, key) != nil
}

func (avl *avlTree) getNode(node *treeNode, key int) *treeNode {
	if node == nil {
		return nil
	}

	if node.key == key {
		return node
	} else if key < node.key {
		return avl.getNode(node.left, key)
	} else {
		return avl.getNode(node.right, key)
	}
}

func max(i, j int) int {
	if i > j {
		return i
	}
	return j
}

func abs(i int) int {
	if i < 0 {
		return -i
	}
	return i
}
