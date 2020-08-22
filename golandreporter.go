package golandreporter

import (
	"fmt"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"github.com/onsi/ginkgo/reporters/stenographer"
	"github.com/onsi/ginkgo/types"
	"os"
	"strings"
	"testing"
	"time"
)

type node struct {
	parent      *node
	description string
	failure     *types.SpecFailure
	time        time.Duration
	testResult  string
	children    map[string]*node
}

var root *node

type GolandReporter struct{}

func NewGolandReporter() reporters.Reporter {
	return GolandReporter{}
}

func NewAutoGolandReporter() reporters.Reporter {
	if strings.Contains(os.Getenv("OLDPWD"), "Goland") {
		return NewGolandReporter()
	} else {
		stenographer := stenographer.New(!config.DefaultReporterConfig.NoColor, config.GinkgoConfig.FlakeAttempts > 1, os.Stdout)
		return reporters.NewDefaultReporter(config.DefaultReporterConfig, stenographer)
	}
}

func (g GolandReporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
	root = &node{nil, "[Top Level]", nil, 0, "", make(map[string]*node)}
}

func (g GolandReporter) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {
	root.print()
}

func (g GolandReporter) SpecWillRun(specSummary *types.SpecSummary) {
	insertNode(root, specSummary.ComponentTexts[1:])
}

func (g GolandReporter) SpecDidComplete(specSummary *types.SpecSummary) {
	if specSummary.Passed() {
		updateResult(root, specSummary, "PASS")
	} else if specSummary.HasFailureState() {
		updateResult(root, specSummary, "FAIL")
		spec, ok := findNode(root, specSummary.ComponentTexts[1:])
		if ok {
			spec.failure = &specSummary.Failure
		}
	} else if specSummary.Skipped() {
		updateResult(root, specSummary, "SKIP")
	} else if specSummary.Pending() {
		updateResult(root, specSummary, "SKIP")
	} else {
		panic("Unknown test output")
	}
}

func updateResult(node *node, specSummary *types.SpecSummary, result string) {
	for i := 1; i < len(specSummary.ComponentTexts); i++ {
		target, ok := findNode(node, specSummary.ComponentTexts[1:i+1])
		target.time = target.time + specSummary.RunTime
		if !ok {
			panic(strings.Join(specSummary.ComponentTexts, "/"))
		}
		if !strings.Contains(target.testResult, "FAIL") {
			target.testResult = result
		}
	}
}

func findNode(node *node, components []string) (*node, bool) {
	if len(components) == 0 {
		return node, true
	}
	component := strings.ReplaceAll(components[0], " ", "_")
	child, ok := node.children[component]
	if !ok {
		return nil, false
	}
	return findNode(child, components[1:])
}

func (g GolandReporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {
	root.print()
}

func (g GolandReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	root.print()
}

func (n node) print() {
	fmt.Printf("=== RUN   %s\n", getSpecName(n))
	for _, node := range n.children {
		node.print()
	}
	if n.failure != nil {
		fmt.Printf("%v\n", n.failure.Location.String())
		fmt.Printf("%s\n\n", n.failure.Message)
		if testing.Verbose() {
			fmt.Printf("%s\n\n", n.failure.Location.FullStackTrace)
		}
	}
	fmt.Printf("--- %s: %s (%.3fs)\n", n.testResult, getSpecName(n), n.time.Seconds())
}

func getSpecName(n node) string {
	if n.description == "[Top Level]" {
		return ""
	}
	name := fmt.Sprintf("%v/%v", getSpecName(*n.parent), n.description)
	return strings.TrimPrefix(name, "/")
}

func insertNode(current *node, components []string) {
	if len(components) < 1 {
		return
	}

	component := strings.ReplaceAll(components[0], " ", "_")
	child, ok := current.children[component]
	if !ok {
		child = &node{current, component, nil, 0,"", make(map[string]*node)}
		current.children[component] = child
	}

	insertNode(child, components[1:])
}