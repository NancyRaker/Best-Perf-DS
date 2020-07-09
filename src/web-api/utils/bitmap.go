/*
@Time : 2019/8/8 17:26 
@Author : yanKoo
@File : bitmap
@Software: GoLand
@Description:
*/
package utils

import (
	"fmt"
	"sync"
)

//位图 @NotThreadSafe
type BitMap struct {
	m sync.Mutex
	bits []byte
	max  int
}

//初始化一个BitMap
//一个byte有8位,可代表8个数字,取余后加1为存放最大数所需的容量
func NewBitMap(max int) *BitMap {
	bits := make([]byte, (max>>3)+1)
	return &BitMap{bits: bits, max: max}
}

//添加一个数字到位图
//计算添加数字在数组中的索引index,一个索引可以存放8个数字
//计算存放到索引下的第几个位置,一共0-7个位置
//原索引下的内容与1左移到指定位置后做或运算
func (b *BitMap) Add(num uint) {
	b.m.Lock()
	defer b.m.Unlock()
	index := num >> 3
	pos := num & 0x07
	b.bits[index] |= 1 << pos
}

//判断一个数字是否在位图
//找到数字所在的位置,然后做与运算
func (b *BitMap) IsExist(num uint) bool {
	b.m.Lock()
	defer b.m.Unlock()
	index := num >> 3
	pos := num & 0x07
	return b.bits[index]&(1<<pos) != 0
}

//删除一个数字在位图
//找到数字所在的位置取反,然后与索引下的数字做与运算
func (b *BitMap) Remove(num uint) {
	b.m.Lock()
	defer b.m.Unlock()
	index := num >> 3
	pos := num & 0x07
	b.bits[index] = b.bits[index] & ^(1 << pos)
}

//位图的最大数字
func (b *BitMap) Max() int {
	return b.max
}

func (b *BitMap) String() string {
	return fmt.Sprint(b.bits)
}