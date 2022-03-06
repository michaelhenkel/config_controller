package graph

import (
	"fmt"
	"sync"
)

// Node a single node that composes the tree
type Node struct {
	name string
	kind string
}

func NewNode(name, kind string) *Node {
	return &Node{
		name: name,
		kind: kind,
	}
}

func (n *Node) String() string {
	return fmt.Sprintf("%v", n.name)
}

func (n *Node) Kind() string {
	return n.kind
}

// ItemGraph the Items graph
type ItemGraph struct {
	nodes []*Node
	edges map[Node][]*Node
	lock  sync.RWMutex
}

// AddNode adds a node to the graph
func (g *ItemGraph) AddNode(n *Node) {
	g.lock.Lock()
	g.nodes = append(g.nodes, n)
	g.lock.Unlock()
}

// AddNode adds a node to the graph
func (g *ItemGraph) GetNode(kind, name string) (*Node, bool) {
	var node *Node
	g.lock.Lock()
	for _, n := range g.nodes {
		if n.kind == kind && n.name == name {
			node = n
		}
	}
	g.lock.Unlock()
	if node != nil {
		return node, true
	}
	return node, false
}

// AddEdge adds an edge to the graph
func (g *ItemGraph) AddEdge(n1, n2 *Node) {
	g.lock.Lock()
	if g.edges == nil {
		g.edges = make(map[Node][]*Node)
	}
	g.edges[*n1] = append(g.edges[*n1], n2)
	g.edges[*n2] = append(g.edges[*n2], n1)
	g.lock.Unlock()
}

// AddEdge adds an edge to the graph
func (g *ItemGraph) String() {
	g.lock.RLock()
	s := ""
	for i := 0; i < len(g.nodes); i++ {
		s += g.nodes[i].String() + " -> "
		near := g.edges[*g.nodes[i]]
		for j := 0; j < len(near); j++ {
			s += near[j].String() + " "
		}
		s += "\n"
	}
	fmt.Println(s)
	g.lock.RUnlock()
}

// Traverse implements the BFS traversing algorithm
func (g *ItemGraph) Traverse(f func(*Node)) {
	g.lock.RLock()
	q := NodeQueue{}
	q.New()
	n := g.nodes[0]
	q.Enqueue(*n)
	visited := make(map[*Node]bool)
	for {
		if q.IsEmpty() {
			break
		}
		fmt.Printf("queue size %d\n", q.Size())
		node := q.Dequeue()
		visited[node] = true
		near := g.edges[*node]

		for i := 0; i < len(near); i++ {
			j := near[i]
			if !visited[j] {
				q.Enqueue(*j)
				visited[j] = true
			}
		}
		if f != nil {
			f(node)
		}
	}
	g.lock.RUnlock()
}

func (g *ItemGraph) TraverseFrom(from *Node, f func(*Node), filterList ...string) {
	var filterMap = make(map[string]struct{})
	var filter bool
	if len(filterList) > 0 {
		filter = true
		for _, filter := range filterList {
			filterMap[filter] = struct{}{}
		}
	}
	g.lock.RLock()
	q := NodeQueue{}
	q.New()
	var n *Node
	for idx, node := range g.nodes {
		if g.nodes[idx].kind == from.kind && g.nodes[idx].name == from.name {
			n = node
		}
	}
	if n == nil {
		return
	}
	q.Enqueue(*n)
	visited := make(map[Node]bool)
	for {
		if q.IsEmpty() {
			break
		}
		node := q.Dequeue()
		visited[*node] = true
		near := g.edges[*node]

		for i := 0; i < len(near); i++ {
			j := near[i]
			ignore := false
			if filter {
				if _, ok := filterMap[j.kind]; !ok {
					ignore = true
				}
			}
			if !visited[*j] && !ignore {
				q.Enqueue(*j)
				visited[*j] = true
			}
		}
		if f != nil {
			f(node)
		}
	}
	g.lock.RUnlock()
}
