package main

import (
	"fmt"

	"github.com/michaelhenkel/config_controller/pkg/graph"
)

func main() {
	g := graph.ItemGraph{}

	/*
		var vnNodes []*graph.Node
		for i := 0; i < 100; i++ {
			n := graph.NewNode(fmt.Sprintf("vn%d", i), "vn")
			g.AddNode(n)
			vnNodes = append(vnNodes, n)
		}

		var vmiNodes []*graph.Node
		for i := 0; i < 60; i++ {
			n := graph.NewNode(fmt.Sprintf("vmi%d", i), "vmi")
			g.AddNode(n)
			vmiNodes = append(vmiNodes, n)
		}

		var vrNodes []*graph.Node
		for i := 0; i < 10; i++ {
			n := graph.NewNode(fmt.Sprintf("vr%d", i), "vr")
			g.AddNode(n)
			vrNodes = append(vrNodes, n)
		}

		var vmNodes []*graph.Node
		for i := 0; i < 30; i++ {
			n := graph.NewNode(fmt.Sprintf("vm%d", i), "vm")
			g.AddNode(n)
			vmNodes = append(vmNodes, n)
		}

		for i := 0; i < len(vrNodes); i++ {
			for j := 0; j < 3; j++ {
				vmNodeIdx := ((i+1)*3 - j) - 1
				fmt.Printf("adding edge from %s/%s to %s/%s\n", vrNodes[i].Kind(), vrNodes[i].String(), vmNodes[vmNodeIdx].Kind(), vmNodes[vmNodeIdx].String())
				g.AddEdge(vrNodes[i], vmNodes[vmNodeIdx])
			}
		}

		for i := 0; i < len(vmiNodes); i++ {
			for j := 0; j < 2; j++ {
				vmiNodeIdx := ((i+1)*2 - j) - 1
				if vmiNodeIdx < len(vmiNodes) {
					fmt.Printf("adding edge from %s/%s to %s/%s\n", vmiNodes[vmiNodeIdx].Kind(), vmiNodes[vmiNodeIdx].String(), vmNodes[i].Kind(), vmNodes[i].String())
					g.AddEdge(vmiNodes[vmiNodeIdx], vmNodes[i])
				}
			}
		}

		for i := 0; i < len(vmiNodes); i++ {
			var vnNode *graph.Node

			res := i % 2
			if res == 1 {
				vnNode = vnNodes[0]
			}

			if res == 0 {
				vnNode = vnNodes[1]
			}

			fmt.Printf("adding edge from %s/%s to %s/%s\n", vmiNodes[i].Kind(), vmiNodes[i].String(), vnNode.Kind(), vnNode.String())
			g.AddEdge(vmiNodes[i], vnNode)
		}

		var nodeList []graph.Node
		g.TraverseFrom(graph.NewNode("vmi19", "vmi"), func(n *graph.Node) {
			if n.Kind() == "vr" {
				nodeList = append(nodeList, *n)
			}
			//fmt.Printf("%s/%s\n", n.Kind(), n.String())
		})
		for _, node := range nodeList {
			fmt.Printf("%s/%s\n", node.Kind(), node.String())
		}
	*/

	vr1 := graph.NewNode("vr1", "vr")
	g.AddNode(vr1)
	vr2 := graph.NewNode("vr2", "vr")
	g.AddNode(vr2)
	vr3 := graph.NewNode("vr3", "vr")
	g.AddNode(vr3)
	vm1 := graph.NewNode("vm1", "vm")
	g.AddNode(vm1)
	vm2 := graph.NewNode("vm2", "vm")
	g.AddNode(vm2)
	vm3 := graph.NewNode("vm3", "vm")
	g.AddNode(vm3)
	vn1 := graph.NewNode("vn1", "vn")
	g.AddNode(vn1)
	vn2 := graph.NewNode("vn2", "vn")
	g.AddNode(vn2)

	iip1 := graph.NewNode("iip1", "iip")
	g.AddNode(iip1)
	//vn2 := graph.NewNode("vn2", "vn")

	vmi1 := graph.NewNode("vmi1", "vmi")
	g.AddNode(vmi1)
	vmi2 := graph.NewNode("vmi2", "vmi")
	g.AddNode(vmi2)

	g.AddEdge(vr1, vm1)
	g.AddEdge(vmi1, vm1)
	g.AddEdge(vmi1, vn1)

	g.AddEdge(iip1, vmi1)

	g.AddEdge(vr2, vm2)
	g.AddEdge(vr3, vm3)
	g.AddEdge(vmi2, vm2)
	g.AddEdge(vmi1, vm3)
	g.AddEdge(vmi2, vn2)

	g.TraverseFrom(vn1, func(n *graph.Node) {
		fmt.Printf("%s/%s\n", n.Kind(), n.String())
	})
	fmt.Println("")
	filterList := []string{"vr", "vn", "vmi", "vm"}
	g.TraverseFrom(vn1, func(n *graph.Node) {
		fmt.Printf("%s/%s\n", n.Kind(), n.String())
	}, filterList...)

	/*

		var nodeList []graph.Node

		g.Traverse(vn1, func(n *graph.Node) {
			if n.Kind() == "vr" {
				nodeList = append(nodeList, *n)
			}
			//fmt.Printf("%s/%s\n", n.Kind(), n.String())
		})

		for _, node := range nodeList {
			fmt.Printf("%s/%s\n", node.Kind(), node.String())
		}
	*/
}
