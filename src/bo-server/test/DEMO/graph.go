/*
@Time : 2019/9/23 11:08 
@Author : yanKoo
@File : graph
@Software: GoLand
@Description:
*/
package main

import (
	"container/list"
	"fmt"
)

func main() {
list.New()
}

type point struct {
	x, y int
}
/*
 1->2->3->4
    2->2->5
 1->4->5->9
*/

func findAvePoint(vertices []*point) []*point {
	var res = []*point{{x: 0, y: 0}}
	// 计算总长
	var lens int
	for i := 1; i < len(vertices); i++ {
		acmulaLens(vertices[i], vertices[i-1], &lens)
	}
	// 起点到终点
	acmulaLens(vertices[0], vertices[len(vertices)-1], &lens)

	fmt.Println(lens, len(vertices))
	aveLen := lens / len(vertices)

	var cur = &point{}
	var next = vertices[1]
	var offset = aveLen
	for i := 1; i < len(vertices); i++ {
		if cur.x == vertices[i].x { // y方向
			if offset <= abs(cur.y, vertices[i].y) {
				next.y = offset
				res = append(res, next)
				i--
				offset = aveLen
				cur = next
				next = &point{x:cur.x}
				continue
			} else {
				offset %= abs(cur.y, vertices[i].y)
				next.y = vertices[i].y
			}
		}

		if cur.y == vertices[i].y { // x方向
			if offset <= abs(cur.x, vertices[i].x) {
				next.x = offset
				res = append(res, next)
				i--
			} else {
				offset %= abs(cur.x, vertices[i].x)
				next.x = vertices[i].x
			}
		}
	}

	return res
}



func abs(i, j int) int {
	if i < j {
		return j - i
	}
	return i - j
}

func acmulaLens(a, b *point, lens *int) {
	if a.x == b.x {
		*lens += abs(a.y, b.y)
	}
	if a.y == b.y {
		*lens += abs(a.x, b.x)
	}
}
