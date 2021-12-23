package main

type ListNode struct {
	Val  int
	Next *ListNode
}

func main() {
	nodeFive := &ListNode{
		Val:  5,
		Next: nil,
	}

	nodeFour := &ListNode{
		Val:  4,
		Next: nodeFive,
	}

	nodeThree := &ListNode{
		Val:  3,
		Next: nodeFour,
	}

	nodeTwo := &ListNode{
		Val:  2,
		Next: nodeThree,
	}

	nodeHead := &ListNode{
		Val:  1,
		Next: nodeTwo,
	}

	reverse(nodeHead)
}

/* 合并两个有序链表
 */

func mergeTwoLists(node1 *ListNode, node2 *ListNode) *ListNode {
	// 虚拟头节点
	var dummy = &ListNode{}
	p := dummy
	var p1 = node1
	var p2 = node2

	for p1 != nil && p2 != nil {
		// 比较 p1 和 p2 两个指针
		if p1.Val > p2.Val {
			p.Next = p2
			p2 = p2.Next
		} else {
			p.Next = p1
			p1 = p1.Next
		}

		p = p.Next
	}

	if p1 != nil {
		p.Next = p1
	}

	if p2 != nil {
		p.Next = p2
	}
	return dummy.Next
}

// 合并 K 个有序链表
func mergeKLists(nodes []ListNode) *ListNode {
	if len(nodes) == 0 {
		return nil
	}

	// 虚拟头结点
	var dummy = &ListNode{}
	_ = dummy
	// 使用 优先级队列（二叉堆） 把链表结点放入一个最小堆

	return nil
}

// 单链表的倒数第K个结点
// 感觉可以用快慢指针,但是应该有更好的办法
func findFormEnd(head *ListNode, k int) *ListNode {
	var p1 = head
	for i := 0; i < k; i++ {
		p1 = p1.Next
	}
	var p2 = head
	for p1 != nil {
		p2 = p2.Next
		p1 = p1.Next
	}
	return p2
}

func removeNthFromEnd(head *ListNode, n int) *ListNode {
	var dummy = &ListNode{Val: -1}
	dummy.Next = head
	var x = findFormEnd(dummy, n+1)
	x.Next = x.Next.Next
	return dummy.Next
}

/*
单向链表
存在一个按升序排列的链表，给你这个链表的头节点 head ，请你删除所有重复的元素，使每个元素 只出现一次 。
*/
func deleteDuplicates(head *ListNode) *ListNode {
	current := head
	for current != nil {
		// 全部删除完再移动到下一个元素
		for current.Next != nil && current.Val == current.Next.Val {
			current.Next = current.Next.Next
		}
		current = current.Next
	}
	return head
}

// 请你删除链表中所有存在数字重复情况的节点，只保留原始链表中 没有重复出现 的数字。
func deleteDuplicatesTwo(head *ListNode) *ListNode {
	current := head
	for current != nil {
		// 全部删除完再移动到下一个元素
		for current.Next != nil && current.Val == current.Next.Val {
			current = current.Next.Next
		}
	}
	return head
}

// 单链表的中点
//输入：[1,2,3,4,5]
//输出：此列表中的结点 3 (序列化形式：[3,4,5])
//返回的结点值为 3 。 (测评系统对该结点序列化表述是 [3,4,5])。
//注意，我们返回了一个 ListNode 类型的对象 ans，这样：
//ans.val = 3, ans.next.val = 4, ans.next.next.val = 5, 以及 ans.next.next.next = NULL.
// 解题思路 快慢指针
//fast/slow 刚开始均指向链表头节点，然后每次快节点走两步，慢指针走一步，直至快指针指向 null，此时慢节点刚好来到链表的下中节点。
func middleNode(head *ListNode) *ListNode {
	var slow = head
	var fast = head
	for fast != nil && fast.Next != nil {
		slow = slow.Next
		fast = fast.Next.Next
	}
	return slow
}

// 判断链表是否包含环
func hasCycle(head *ListNode) bool {
	var slow = head
	var fast = head
	for fast != nil && fast.Next != nil {
		slow = slow.Next
		fast = fast.Next.Next

		if slow == fast {
			return true
		}
	}
	return false
}

// 翻转链表
// 将第一个放在最后面
func reverseList(head *ListNode) *ListNode {
	if head == nil || head.Next == nil {
		return head
	}

	var node = reverseList(head.Next)
	head.Next.Next = head
	head.Next = nil
	return node
}

// 求两条链表的相交节点
// a1-> a2 -> c1 ->c2
// b1-> b2 -> b3 -> c1 ->c2

// 相交链表后
// a1-> a2 -> c1 -> c2 -> b1 -> b2 -> b3 -> c1
// b1-> b2 -> b3 -> c1 -> c2 -> a1 -> a2 -> c1
// 这里能看到通过 链表A 和链表B的相交,能同时达到交点 C1
func getIntersectionNode(node1 *ListNode, node2 *ListNode) *ListNode {
	var p1 = node1
	var p2 = node2
	for p1 != p2 {
		if p1 == nil {
			p1 = node2
		} else {
			p1 = p1.Next
		}

		if p2 == nil {
			p2 = node1
		} else {
			p2 = p2.Next
		}

	}
	return p1
}

// 给你一个链表的头节点 head 和一个整数 val ，请你删除链表中所有满足 Node.val == val 的节点，并返回 新的头节点 。
func removeElements(head *ListNode, val int) *ListNode {
	// 创建一个虚拟结点
	dummyHead := &ListNode{Next: head}
	for tmp := dummyHead; tmp.Next != nil; {
		if tmp.Next.Val == val {
			tmp.Next = tmp.Next.Next
		} else {
			tmp = tmp.Next
		}
	}
	return dummyHead
}

// 翻转链表II
//给你单链表的头指针 head 和两个整数 left 和 right ，其中 left <= right 。请你反转从位置 left 到位置 right 的链表节点，返回 反转后的链表 。
//
//思路:先保存 left 前面的 head
//然后保存 right 后面的
//中间部分变成一个单独的链表，然后直接进行反转
func reverseBetween(head *ListNode, left, right int) *ListNode {
	// 因为头节点有可能发生变化，使用虚拟头节点可以避免复杂的分类讨论
	dummyNode := &ListNode{Val: -1}
	dummyNode.Next = head

	pre := dummyNode
	// 第 1 步：从虚拟头节点走 left - 1 步，来到 left 节点的前一个节点
	// 建议写在 for 循环里，语义清晰
	for i := 0; i < left-1; i++ {
		pre = pre.Next
	}

	// 第 2 步：从 pre 再走 right - left + 1 步，来到 right 节点
	rightNode := pre
	for i := 0; i < right-left+1; i++ {
		rightNode = rightNode.Next
	}

	// 第 3 步：切断出一个子链表（截取链表）
	leftNode := pre.Next
	curr := rightNode.Next

	// 注意：切断链接
	pre.Next = nil
	rightNode.Next = nil

	// 第 4 步：同第 206 题，反转链表的子区间
	reverseLinkedList(leftNode)

	// 第 5 步：接回到原来的链表中
	pre.Next = rightNode
	leftNode.Next = curr
	return dummyNode.Next
}

func reverseLinkedList(head *ListNode) {
	var pre *ListNode
	cur := head
	for cur != nil {
		next := cur.Next
		cur.Next = pre
		pre = cur
		cur = next
	}
}

// 递归翻转链表
func reverse(head *ListNode) *ListNode {
	if head.Next == nil {
		return head
	}
	last := reverse(head.Next)
	head.Next.Next = head
	head.Next = nil
	return last
}

var successor = &ListNode{}

func reverseN(head *ListNode, n int) *ListNode {
	if n == 1 {
		successor = head.Next
		return head
	}

	var last = reverseN(head.Next, n-1)
	head.Next.Next = head
	head.Next = successor
	return last
}

func NewReverseBetween(head *ListNode, n int, m int) *ListNode {
	if m == 1 {
		return reverseN(head, n)
	}
	head.Next = reverseBetween(head.Next, m-1, n-1)
	return head
}
