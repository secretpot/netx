package tree

import (
	"sync"

	"github.com/secretpot/netx/ipv4"

	jmath "github.com/secretpot/jbutil/math"
	jtype "github.com/secretpot/jbutil/types"
)

type NetTree struct {
	rootNode     *HopNode
	Destinations jtype.Set
	Nodes        map[string]*HopNode
}

func (this *NetTree) Localhost() string {
	return this.rootNode.IP()
}
func (this *NetTree) Root() *HopNode {
	return this.rootNode
}
func (this *NetTree) AllDestinations() []string {
	dests := []string{}
	for k := range this.Destinations {
		dests = append(dests, k.(string))
	}
	return dests
}
func (this *NetTree) GetHopNode(ip string) *HopNode {
	return this.Nodes[ip]
}

func (this *NetTree) HowToGo(ip string) []*HopNode {
	return this.Root().FindPath(ip)
}
func (this *NetTree) AllCanGo() map[string][]*HopNode {
	return this.Root().AllPath()
}
func (this *NetTree) DFS() []*HopNode {
	return this.Root().DFS()
}
func (this *NetTree) BFS() []*HopNode {
	return this.Root().BFS()
}

func (this *NetTree) Absorb(subtree *HopNode) {
	/* 吸收给定的子树: 添加子树可以通往的dest, 并将所有节点加入查找集 */
	nodeDFS := subtree.DFS()
	// 仅添加子树根节点可以通往的dest, 子树非根节点可以通往的其他dest无需关注(子树根节点无法到达的dest, 其上级节点无需了解)
	// 若无根节点同级或更上级的节点可通往该dest, 说明该dest没有可达路径, 无需关注
	// 若有, 则会在处理全局根节点或其他子树时添加
	for domain := range subtree.next {
		this.Destinations.Add(domain)
	}
	for _, node := range nodeDFS {
		// 对于待加入的node, 树中不存在该node则加入, 存在则合并下一跳信息, 不保存未知节点
		if node.IP() == "*" {
		} else if _, ok := this.Nodes[node.IP()]; !ok {
			this.Nodes[node.IP()] = node
		} else {
			this.Nodes[node.IP()].AppendNexts(node.AllNext())
		}
	}
}

func (this *NetTree) InsertFullPath(treeData map[string][]ipv4.EchoSummary) *NetTree {
	/* 根据{dest: path}插入完整路径 */
	for domain, path := range treeData {
		node := this.Root()
		ttl := 1
		for _, data := range path {
			var nextNode *HopNode
			var nextNodeDests jtype.Set
			if data.TraceIP() == "*" {
				nextNode = NewLeafHopNode("*", make(Attrs))
			} else {
				nextNode = this.Nodes[data.TraceIP()]
				nextNode = jmath.If(nextNode != nil, nextNode, NewLeafHopNode(data.TraceIP(), make(Attrs))).(*HopNode)
			}
			nextNode.SetAttr("ttl", ttl)
			nextNode.SetAttr("destinations", jtype.Set{})
			nextNodeDests = nextNode.GetAttr("destinations").(jtype.Set)
			nextNodeDests.Add(domain)
			node.AddNext(domain, nextNode)
			this.Absorb(nextNode) // 一定要在此时吸收新节点, 否则下一次添加同一个节点时无法查找到该节点, 会导致重复创建
			node = nextNode
			ttl = ttl + 1
		}
		this.Destinations.Add(domain) // 一定要在此时手动添加终点, 因为每个新节点在吸收时都没有下一跳信息
	}

	return this
}

func (this *NetTree) Copy() *NetTree {
	copyTree := &NetTree{
		this.rootNode.Duplicate(),
		this.Destinations.Copy(), // 此处将终点信息都吸收了, 若后续没有对应路径则会寻路失败
		map[string]*HopNode{},
	}
	copyTree.Absorb(copyTree.rootNode)

	for dest, path := range this.AllCanGo() {
		node := copyTree.rootNode
		for _, point := range path[1:] {
			nextHop := copyTree.Nodes[point.IP()]
			nextHop = jmath.If(nextHop != nil, nextHop, point.Duplicate()).(*HopNode)
			node.AddNext(dest, nextHop)
			copyTree.Absorb(nextHop)
			node = nextHop
		}
	}

	// 确保一下copy之后tree的正确性
	allNexts := copyTree.rootNode.AllNext()
	for _, dest := range copyTree.AllDestinations() {
		if _, ok := allNexts[dest]; !ok {
			copyTree.Destinations.Del(dest)
		}
	}

	return copyTree
}

func NewNetTree() *NetTree {
	root := &HopNode{
		ipv4.Localhost(),
		make(Attrs),
		make(map[string]*HopNode),
	}
	return &NetTree{
		root,
		jtype.Set{},
		map[string]*HopNode{root.IP(): root},
	}
}

func BuildNetTreeFrom(domains ...string) *NetTree {
	routerTree := NewNetTree()

	wg := new(sync.WaitGroup)
	dataChan := make(chan []interface{}, len(domains))
	trobeData := make(map[string][]ipv4.EchoSummary)

	wg.Add(len(domains))
	for _, dm := range domains {
		go func(domain string) {
			defer wg.Done()
			dataChan <- []interface{}{domain, ipv4.SimpleTrobeWithEnsurance(domain, 1, 1.0, 3)} // SimpleTrobe is enough
		}(dm)
	}
	wg.Wait()
	close(dataChan)

	for data := range dataChan {
		trobeData[data[0].(string)] = data[1].([]ipv4.EchoSummary)
	}
	return routerTree.InsertFullPath(trobeData)
}
