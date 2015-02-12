package main

import (
	"encoding/xml"
	"flag"
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

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: %s [OPTIONS] FILE

FILE is the path to an exported VisualVM call tree in XML format.
Exported subtree call stacks work in addition to the full call stack,
and will process more quickly.

OPTIONS are:
`, os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	var rootName string
	var numResults int
	var ignoreNames stringslice
	flag.StringVar(&rootName, "root", "", "Treat first matching function name as root node")
	flag.IntVar(&numResults, "n", 50, "Report first n results")
	flag.Var(&ignoreNames, "ignore", "Ignore matching function names (may specify multiple)")
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	filename := flag.Arg(0)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	v := struct {
		Tree Node `xml:"tree"`
	}{}
	xml.Unmarshal(data, &v)

	root := v.Tree.Nodes[0]

	nodes := []*Node{root}
	for i := 0; i < len(nodes); i++ {
		if strings.Contains(nodes[i].Name, rootName) {
			root = nodes[i]
			break
		}
		nodes = append(nodes, nodes[i].Nodes...)
	}

	nodes = []*Node{root}
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
outer:
	for _, pair := range pairs {
		for _, ignoreName := range ignoreNames {
			if strings.Contains(pair.Name, ignoreName) {
				continue outer
			}
		}
		fmt.Fprintf(w, "%s\t%f\n", pair.Name, pair.RelTimeCPU)
		nPrinted++
		if nPrinted >= numResults {
			break
		}
	}
	w.Flush()
}

type stringslice []string

func (ss *stringslice) Set(s string) error {
	*ss = append(*ss, s)
	return nil
}

func (ss stringslice) String() string {
	return fmt.Sprint([]string(ss))
}

type ByRelTimeCPU []NameRelTimeCPU

func (a ByRelTimeCPU) Len() int           { return len(a) }
func (a ByRelTimeCPU) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByRelTimeCPU) Less(i, j int) bool { return a[i].RelTimeCPU < a[j].RelTimeCPU }
