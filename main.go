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
}

func (n *Node) ComputeRelTimeCPU(totalTimeCPU int) {
	n.RelTimeCPU = float32(n.TimeCPU) / float32(totalTimeCPU)
}

func (n *Node) ComputeRelTimeCPURecursive(totalTimeCPU int) {
	n.ComputeRelTimeCPU(totalTimeCPU)
	for _, node := range n.Nodes {
		node.ComputeRelTimeCPURecursive(totalTimeCPU)
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

	root.ComputeRelTimeCPURecursive(root.TimeCPU)

	os.Exit(0)

	nodes := []*Node{v.Tree.Nodes[0]}
	i := 0
	for i < len(nodes) {
		nodes = append(nodes, nodes[i].Nodes...)
		i++
	}

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

	totalTimeCPU := v.Tree.Nodes[0].TimeCPU
	pairs := []NameTimePair{}
	for name, timeCPU := range nodesTimeCPU {
		pair := NameTimePair{name, float32(timeCPU) / float32(totalTimeCPU)}
		pairs = append(pairs, pair)
	}
	sort.Sort(sort.Reverse(ByRelTimeCPU(pairs)))

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, ' ', 0)

	for i := 0; i < 100; i++ {
		fmt.Fprintf(w, "%s\t%f\n", pairs[i].Name, pairs[i].RelTimeCPU)
	}
	w.Flush()

	pct := float32(nodesTimeCPU["ml_lib.geo_ip$ip_to_location.invoke()"]) /
		float32(nodesTimeCPU["haggler.handler$bid.invoke()"])
	fmt.Println(pct)
}

type ByRelTimeCPU []NameTimePair

func (a ByRelTimeCPU) Len() int           { return len(a) }
func (a ByRelTimeCPU) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByRelTimeCPU) Less(i, j int) bool { return a[i].RelTimeCPU < a[j].RelTimeCPU }
