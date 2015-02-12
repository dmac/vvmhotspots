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
	Name    string
	Parent  string
	TimeCPU int     `xml:"Time-CPU"`
	Nodes   []*Node `xml:"node"`
}

type NameRelTimeCPU struct {
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
	nodes := []*Node{root}
	for i := 0; i < len(nodes); i++ {
		nodes = append(nodes, nodes[i].Nodes...)
	}

	nodesTimeCPU := map[string]int{}
	for _, node := range nodes {
		// TODO(dmac) This prevents overloaded functions from being double-counted. But, does it also
		// mask time spent in recursive functions?
		if node.Name == node.Parent {
			continue
		}
		if _, found := nodesTimeCPU[node.Name]; !found {
			nodesTimeCPU[node.Name] = 0
		}
		nodesTimeCPU[node.Name] += node.TimeCPU
	}

	pairs := []NameRelTimeCPU{}
	for name, timeCPU := range nodesTimeCPU {
		pair := NameRelTimeCPU{name, float32(timeCPU) / float32(root.TimeCPU)}
		pairs = append(pairs, pair)
	}

	sort.Sort(sort.Reverse(ByRelTimeCPU(pairs)))

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 1, ' ', 0)
	nPrinted := 0
	nToPrint := 50
	for _, pair := range pairs {
		if strings.HasPrefix(pair.Name, "clojure.lang") ||
			strings.HasPrefix(pair.Name, "clojure.core") ||
			pair.Name == "Self time" {
			continue
		}
		fmt.Fprintf(w, "%s\t%f\n", pair.Name, pair.RelTimeCPU)
		nPrinted++
		if nPrinted >= nToPrint {
			break
		}
	}
	w.Flush()
}

type ByRelTimeCPU []NameRelTimeCPU

func (a ByRelTimeCPU) Len() int           { return len(a) }
func (a ByRelTimeCPU) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByRelTimeCPU) Less(i, j int) bool { return a[i].RelTimeCPU < a[j].RelTimeCPU }
