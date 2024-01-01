package memdb

import (
	"bytes"
)

type List struct {
	Head *ListNode
	Tail *ListNode
	Len  int
}

type ListNode struct {
	Prev *ListNode
	Next *ListNode
	Val  []byte
}

func NewList() *List {
	// dummy node
	head := &ListNode{}
	tail := &ListNode{}
	head.Next = tail
	tail.Prev = head
	return &List{Head: head, Tail: tail, Len: 0}
}

func (l *List) Index(index int) *ListNode {
	var cur *ListNode
	if index < 0 {
		if -index > l.Len {
			return nil
		}
		cur = l.Tail.Prev
		for i := 0; i < -index-1; i++ {
			cur = cur.Prev
		}
	} else {
		if index >= l.Len {
			return nil
		}
		cur = l.Head.Next
		// 这里不包括index的原因是因为以及access了一次node
		for i := 0; i < index; i++ {
			cur = cur.Next
		}
	}
	return cur
}

func (l *List) Pos(val []byte) int {
	pos := 0
	for cur := l.Head.Next; cur != l.Tail; cur = cur.Next {
		if bytes.Equal(cur.Val, val) {
			return pos
		}
		pos++
	}
	return -1
}

func (l *List) InsertBefore(val []byte, pivot []byte) bool {
	var ok bool
	for cur := l.Head.Next; cur != l.Tail; cur = cur.Next {
		if bytes.Equal(cur.Val, pivot) {
			ok = true
			node := &ListNode{Prev: cur.Prev, Next: cur, Val: val}
			cur.Prev = node
			node.Prev.Next = node
			l.Len++
		}
	}
	return ok
}

func (l *List) InsertAfter(val []byte, pivot []byte) bool {
	var ok bool
	for cur := l.Head.Next; cur != l.Tail; cur = cur.Next {
		if bytes.Equal(cur.Val, pivot) {
			ok = true
			node := &ListNode{Prev: cur, Next: cur.Next, Val: val}
			cur.Next = node
			node.Next.Prev = node
			l.Len++
		}
	}
	return ok
}

func (l *List) LPop() *ListNode {
	if l.Len == 0 {
		return nil
	}
	node := l.Head.Next

	l.Head.Next = node.Next
	node.Next.Prev = l.Head
	l.Len--
	return node
}

func (l *List) RPop() *ListNode {
	if l.Len == 0 {
		return nil
	}
	node := l.Tail.Prev

	l.Tail.Prev = node.Prev
	node.Prev.Next = l.Tail
	l.Len--
	return node
}

func (l *List) LPush(val []byte) {
	node := &ListNode{Next: l.Head.Next, Prev: l.Head, Val: val}
	l.Head.Next = node
	node.Next.Prev = node
	l.Len++
}

func (l *List) RPush(val []byte) {
	node := &ListNode{Next: l.Tail, Prev: l.Tail.Prev, Val: val}
	l.Tail.Prev = node
	node.Prev.Next = node
	l.Len++
}

func (l *List) Range(start, end int) [][]byte {
	if start < 0 {
		start = l.Len + start
	}
	if end < 0 {
		end = l.Len + end
	}

	if start > end || start >= l.Len || end < 0 {
		return nil
	}

	if end >= l.Len {
		end = l.Len - 1
	}

	if start < 0 {
		start = 0
	}

	res := make([][]byte, 0, end-start+1)
	cur := l.Head
	for i := 0; i <= end; i++ {
		cur = cur.Next
		if i >= start {
			res = append(res, cur.Val)
		}
	}
	return res
}

/*
count > 0: Remove elements equal to element moving from head to tail.
count < 0: Remove elements equal to element moving from tail to head.
count = 0: Remove all elements equal to element.
*/
func (l *List) Remove(val []byte, count int) int {
	if l.Len == 0 {
		return 0
	}

	if count == 0 {
		count = l.Len
	}

	removed := 0
	if count > 0 {
		for cur := l.Head.Next; cur != l.Tail && removed < count; {
			if bytes.Equal(cur.Val, val) {
				cur.Next.Prev = cur.Prev
				cur.Prev.Next = cur.Next
				l.Len--
				removed++
			}
			cur = cur.Next
		}
	} else {
		for cur := l.Tail.Prev; cur != l.Head && removed < -count; {
			if bytes.Equal(cur.Val, val) {
				cur.Next.Prev = cur.Prev
				cur.Prev.Next = cur.Next
				l.Len--
				removed++
			}
			cur = cur.Prev
		}
	}

	return removed
}

func (l *List) Set(val []byte, index int) bool {
	node := l.Index(index)
	if node == nil {
		return false
	}
	node.Val = val
	return true
}

func (l *List) Trim(start, end int) {
	if start < 0 {
		start = l.Len + start
	}
	if end < 0 {
		end = l.Len + end
	}

	if start > end || start >= l.Len || end < 0 {
		l.Clear()
		return
	}

	if end >= l.Len {
		end = l.Len - 1
	}

	if start < 0 {
		start = 0
	}

	var startNode, endNode *ListNode
	pos := 0
	for cur := l.Head.Next; cur != l.Tail; cur = cur.Next {
		if pos == start {
			startNode = cur
		}
		if pos == end {
			endNode = cur
			break
		}
		pos++
	}

	l.Head.Next = startNode
	startNode.Prev = l.Head
	l.Tail.Prev = endNode
	endNode.Next = l.Tail
	l.Len = end - start + 1
}

func (l *List) Clear() {
	if l.Len == 0 {
		return
	}

	l.Head.Next = l.Tail
	l.Tail.Prev = l.Head
	l.Len = 0
}
