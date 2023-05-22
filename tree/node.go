package tree

import (
	"fmt"
	"strings"
)

type HopNode struct {
	ip    string
	attrs Attrs
	next  map[string]*HopNode
}

func (this *HopNode) IP() string {
	return this.ip
}
func (this *HopNode) GetAttr(key string) interface{} {
	return this.attrs[key]
}
func (this *HopNode) SetAttr(key string, value interface{}) {
	if _, ok := this.attrs[key]; !ok {
		this.attrs[key] = value
	}
}
func (this *HopNode) ForceSetAttr(key string, value interface{}) {
	this.attrs[key] = value
}
func (this *HopNode) RemoveAttr(keys ...string) {
	for _, k := range keys {
		delete(this.attrs, k)
	}
}
func (this *HopNode) NextHopTo(target string) *HopNode {
	return this.next[target]
}
func (this *HopNode) AllNext() map[string]*HopNode {
	nexts := make(map[string]*HopNode)
	for k, v := range this.next {
		nexts[k] = v
	}
	return nexts
}

func (this *HopNode) String() string {
	return fmt.Sprintf("%v(%v)", this.ip, this.attrs)
}
func (this *HopNode) AddNextLeaf(domain string, ip string, attrs Attrs) *HopNode {
	/* 根据给定的ip创建一个新的初始节点作为this节点去往domain的下一跳, 若已存在下一跳则不进行操作 */
	var newLeaf *HopNode
	if _, ok := this.next[domain]; !ok {
		newLeaf = NewLeafHopNode(ip, attrs)
		this.next[domain] = newLeaf
	}
	return newLeaf
}
func (this *HopNode) AddNext(domain string, node *HopNode) *HopNode {
	/* 将node设置为this节点去往domain的下一跳, 若已存在下一跳则不进行操作 */
	if _, ok := this.next[domain]; !ok {
		this.next[domain] = node
		return node
	}
	return nil
}
func (this *HopNode) AppendNexts(next map[string]*HopNode) map[string]*HopNode {
	/* 将给定的下一跳集合追加进this节点, 跳过已有下一跳的domain */
	r := make(map[string]*HopNode)
	for domain, node := range next {
		if appended := this.AddNext(domain, node); appended != nil {
			r[domain] = appended
		}
	}
	return r
}
func (this *HopNode) FindPath(target string) []*HopNode {
	/* 寻找去往ip的路径 */
	result := []*HopNode{}
	for node := this; node != nil; node = node.next[target] {
		result = append(result, node)
	}
	// result = util.If(len(result) == 1, []*HopNode{}, result).([]*HopNode)
	return result
}
func (this *HopNode) AllPath() map[string][]*HopNode {
	/* 寻找从this出发,所有去往可以到达的终点的路径 */
	result := make(map[string][]*HopNode)
	for domain := range this.next {
		result[domain] = this.FindPath(domain)
	}
	return result
}
func (this *HopNode) Duplicate() *HopNode {
	return NewLeafHopNode(this.IP(), this.attrs.Copy())
}
func (this *HopNode) Regenerate() map[string]*HopNode {
	children := map[string]*HopNode{}
	for dest, child := range this.AllNext() {
		children[dest] = child.Duplicate()
	}
	return children
}
func (this *HopNode) dfs() []*HopNode {
	// RECURSIVELY IMPLEMENT
	// result := []*HopNode{}
	// for _, node := range this.next {
	// 	if node.GetAttr("visit") != true {
	// 		result = append(result, node)
	// 		node.SetAttr("visit", true)
	// 	}
	// 	if len(node.next) > 0 {
	// 		result = append(result, node.dfs()...)
	// 	}
	// }
	result := []*HopNode{}
	stack := []*HopNode{this}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if node.GetAttr("visit") != true {
			result = append(result, node)
			node.SetAttr("visit", true)
		}
		for _, subNode := range node.next {
			stack = append(stack, subNode)
		}
	}
	return result
}
func (this *HopNode) DFS() []*HopNode {
	result := this.dfs()
	for _, node := range result {
		node.RemoveAttr("visit")
	}
	// RECURSIVELY IMPLEMENT
	// result = append([]*HopNode{this}, result...)
	return result
}
func (this *HopNode) bfs() []*HopNode {
	result := []*HopNode{}
	queue := []*HopNode{this}
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		if node.GetAttr("visit") != true {
			result = append(result, node)
			node.SetAttr("visit", true)
		}
		for _, subNode := range node.next {
			queue = append(queue, subNode)
		}
	}
	return result
}
func (this *HopNode) BFS() []*HopNode {
	result := this.bfs()
	for _, node := range result {
		node.RemoveAttr("visit")
	}
	return result
}

func NewHopNode(ip string, attrs Attrs, next map[string]*HopNode) *HopNode {
	return &HopNode{
		ip,
		attrs,
		next,
	}
}
func NewLeafHopNode(ip string, attrs Attrs) *HopNode {
	return NewHopNode(ip, attrs, make(map[string]*HopNode))
}

func PathString(path []*HopNode) string {
	nodes := []string{}
	for _, node := range path {
		nodes = append(nodes, node.IP())
	}
	return strings.Join(nodes, " -> ")
}
