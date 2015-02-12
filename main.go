package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

type Node struct {
	Name       string
	TimeCPU    int `xml:"Time-CPU"`
	RelTimeCPU float32
	Nodes      []*Node `xml:"node"`
	Depth      int
}

func (n *Node) ComputeRelTimeCPU(totalTimeCPU int) {
	n.RelTimeCPU = float32(n.TimeCPU) / float32(totalTimeCPU)
}

func (n *Node) ComputeRelTimeCPURecursive(totalTimeCPU, depth int) {
	n.ComputeRelTimeCPU(totalTimeCPU)
	n.Depth = depth
	for _, node := range n.Nodes {
		node.ComputeRelTimeCPURecursive(totalTimeCPU, depth+1)
	}
}

type NameTimePair struct {
	Name       string
	RelTimeCPU float32
}

func main() {
	data, err := ioutil.ReadFile("/Users/dmac/VisualVM Snapshots/csv/geoip1.bid-subtree.xml")
	if err != nil {
		panic(err)
	}
	v := struct {
		Tree Node `xml:"tree"`
	}{}
	xml.Unmarshal(data, &v)
	root := v.Tree.Nodes[0]

	root.ComputeRelTimeCPURecursive(root.TimeCPU, 0)

	nodes := []*Node{}
	dft := []*Node{root}
	var node *Node
	for len(dft) > 0 {
		node, dft = dft[len(dft)-1], dft[:len(dft)-1]
		sort.Sort(ByRelTimeCPU(node.Nodes))
		nodes = append(nodes, node)
		dft = append(dft, node.Nodes...)
	}

	//for _, node := range nodes {
	//if strings.hasprefix(node.name, "clojure.lang") ||
	//strings.hasprefix(node.name, "clojure.core") ||
	//node.name == "self time" {
	//continue
	//}
	//for i := 0; i < node.Depth; i++ {
	//fmt.Print("  ")
	//}
	//fmt.Println(node.Name, node.RelTimeCPU)
	//}

	//os.Exit(0)

	nodesTimeCPU := map[string]int{}
	for _, node := range nodes {
		if strings.HasPrefix(node.Name, "clojure.lang") ||
			strings.HasPrefix(node.Name, "clojure.core") ||
			node.Name == "Self time" {
			continue
		}
		if _, found := nodesTimeCPU[node.Name]; !found {
			nodesTimeCPU[node.Name] = 0
		}
		nodesTimeCPU[node.Name] += node.TimeCPU
	}

	//totalTimeCPU := v.Tree.Nodes[0].TimeCPU
	//pairs := []NameTimePair{}
	//for name, timeCPU := range nodesTimeCPU {
	//pair := NameTimePair{name, float32(timeCPU) / float32(totalTimeCPU)}
	//pairs = append(pairs, pair)
	//}

	//sort.Sort(sort.Reverse(ByRelTimeCPU(pairs)))

	sort.Sort(sort.Reverse(ByRelTimeCPU(nodes)))

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, ' ', 0)

	nPrinted := 0
	for _, node := range nodes {
		if strings.HasPrefix(node.Name, "clojure.lang") ||
			strings.HasPrefix(node.Name, "clojure.core") ||
			node.Name == "self time" {
			continue
		}
		fmt.Fprintf(w, "%s\t%f\n", node.Name, node.RelTimeCPU)
		nPrinted++
		if nPrinted >= 100 {
			break
		}
	}
	w.Flush()

	pct := float32(nodesTimeCPU["ml_lib.geo_ip$ip_to_location.invoke()"]) /
		float32(nodesTimeCPU["haggler.handler$bid.invoke()"])
	fmt.Println(pct)
}

type ByRelTimeCPU []*Node

func (a ByRelTimeCPU) Len() int           { return len(a) }
func (a ByRelTimeCPU) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByRelTimeCPU) Less(i, j int) bool { return a[i].RelTimeCPU < a[j].RelTimeCPU }
