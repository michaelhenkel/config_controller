package graph

import (
	"sync"
)

// Node a single node that composes the tree
type Node struct {
	Name      string
	Namespace string
	Kind      string
}

// ItemGraph the Items graph
type ItemGraph struct {
	nodes map[Node]*Node
	edges map[Node][]*Node
	lock  sync.RWMutex
}

// AddNode adds a node to the graph
func (g *ItemGraph) AddNode(n Node) {
	g.lock.Lock()
	if g.nodes == nil {
		g.nodes = make(map[Node]*Node)
	}
	g.nodes[n] = &n
	g.lock.Unlock()
}

// AddNode adds a node to the graph
func (g *ItemGraph) GetNode(node Node) (*Node, bool) {
	g.lock.Lock()
	n, ok := g.nodes[node]
	g.lock.Unlock()
	return n, ok
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

func (g *ItemGraph) TraverseFrom(from Node, to *Node, f func(*Node), filterList ...string) {
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
	n, ok := g.nodes[from]
	if !ok {
		g.lock.RUnlock()
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
		if node.Kind != to.Kind {
			near := g.edges[*node]
			for i := 0; i < len(near); i++ {
				j := near[i]
				ignore := false
				if filter {
					if _, ok := filterMap[j.Kind]; !ok {
						ignore = true
					}
				}
				if !visited[*j] && !ignore {
					q.Enqueue(*j)
					visited[*j] = true
				}
			}
		}
		if f != nil {
			f(node)
		}
	}
	g.lock.RUnlock()
}
