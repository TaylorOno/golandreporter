package golandreporter

import (
	"fmt"
	"github.com/onsi/ginkgo/types"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGolandReporter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithCustomReporters(t, "Golandreporter Suite", []Reporter{NewGolandReporter()})
}

var _ = Describe("GolandReporter", func(){
	var root *node

	BeforeEach(func(){
		root = &node{nil, "[Top Level]", nil, 0, "", []*node{}}
	})

	Context("insertNode", func(){
		It("inserts describe node", func(){
			insertNode(root, []string{"DescribeBlock"})
			Expect(len(root.children)).To(Equal(1))
			Expect(root.children[0].description).To(Equal("DescribeBlock"))
		})

		It("does not inserts duplicate elements", func(){
			insertNode(root, []string{"DescribeBlock"})
			insertNode(root, []string{"DescribeBlock"})
			Expect(len(root.children)).To(Equal(1))
		})

		It("inserts context blocks elements", func(){
			insertNode(root, []string{"DescribeBlock", "ContextBlock"})
			Expect(len(root.children)).To(Equal(1))
			Expect(root.children[0].description).To(Equal("DescribeBlock"))
			Expect(len(root.children[0].children)).To(Equal(1))
			Expect(root.children[0].children[0].description).To(Equal("ContextBlock"))
		})

		It("inserts multiple context blocks elements", func(){
			insertNode(root, []string{"DescribeBlock", "ContextBlock1"})
			insertNode(root, []string{"DescribeBlock", "ContextBlock2"})
			Expect(len(root.children)).To(Equal(1))
			Expect(root.children[0].description).To(Equal("DescribeBlock"))
			Expect(len(root.children[0].children)).To(Equal(2))
			Expect(root.children[0].children[0].description).To(Equal("ContextBlock1"))
			Expect(root.children[0].children[1].description).To(Equal("ContextBlock2"))
		})
	})

	Context("getSpecName", func(){
		It("returns a full spec name given a node", func(){
			insertNode(root, []string{"DescribeBlock", "ContextBlock", "SpecBlock"})
			describeBlock := *root.children[0]
			contextBlock := *root.children[0].children[0]
			specBlock := *root.children[0].children[0].children[0]
			Expect(getSpecName(describeBlock)).To(Equal("DescribeBlock"))
			Expect(getSpecName(contextBlock)).To(Equal("DescribeBlock/ContextBlock"))
			Expect(getSpecName(specBlock)).To(Equal("DescribeBlock/ContextBlock/SpecBlock"))
		})
	})

	Context("findNode", func(){
		It("returns true and node if found", func(){
			insertNode(root, []string{"DescribeBlock", "ContextBlock", "SpecBlock"})
			node, ok := findNode(root, []string{"DescribeBlock", "ContextBlock", "SpecBlock"})
			Expect(ok).To(BeTrue())
			Expect(node.description).To(Equal("SpecBlock"))
		})

		It("returns false and nil spec node is not found", func(){
			insertNode(root, []string{"DescribeBlock", "ContextBlock", "SpecBlock"})
			node, ok := findNode(root, []string{"DescribeBlock", "ContextBlock", "MissingSpecBlock"})
			Expect(ok).To(BeFalse())
			Expect(node).To(BeNil())
		})

		It("returns false and nil context node is not found", func(){
			insertNode(root, []string{"DescribeBlock", "ContextBlock", "SpecBlock"})
			node, ok := findNode(root, []string{"DescribeBlock", "MissingContextBlock", "SpecBlock"})
			Expect(ok).To(BeFalse())
			Expect(node).To(BeNil())
		})
	})

	Context("updateResult", func(){
		It("inserts PASS at all levels", func(){
			insertNode(root, []string{"DescribeBlock", "ContextBlock", "SpecBlock"})
			summary := &types.SpecSummary{ComponentTexts:[]string{"[Top_Level]", "DescribeBlock", "ContextBlock", "SpecBlock"}, RunTime:1}
			updateResult(root, summary, "PASS")
			Expect(root.children[0].testResult).To(Equal("PASS"))
			Expect(root.children[0].children[0].testResult).To(Equal("PASS"))
			Expect(root.children[0].children[0].children[0].testResult).To(Equal("PASS"))
		})

		It("inserts FAIL at all levels", func(){
			insertNode(root, []string{"DescribeBlock", "ContextBlock", "SpecBlock"})
			summary := &types.SpecSummary{ComponentTexts:[]string{"[Top_Level]", "DescribeBlock", "ContextBlock", "SpecBlock"}, RunTime:1}
			updateResult(root, summary, "FAIL")
			Expect(root.children[0].testResult).To(Equal("FAIL"))
			Expect(root.children[0].children[0].testResult).To(Equal("FAIL"))
			Expect(root.children[0].children[0].children[0].testResult).To(Equal("FAIL"))
		})

		It("doesn't update parent nodes after failed child", func(){
			insertNode(root, []string{"DescribeBlock", "ContextBlock", "SpecBlock1"})
			insertNode(root, []string{"DescribeBlock", "ContextBlock", "SpecBlock2"})
			testSummary1 := &types.SpecSummary{ComponentTexts:[]string{"[Top_Level]", "DescribeBlock", "ContextBlock", "SpecBlock1"}, RunTime:1}
			updateResult(root, testSummary1, "FAIL")
			testSummary2 := &types.SpecSummary{ComponentTexts:[]string{"[Top_Level]", "DescribeBlock", "ContextBlock", "SpecBlock2"}, RunTime:1}
			updateResult(root, testSummary2, "PASS")
			Expect(root.children[0].testResult).To(Equal("FAIL"))
			Expect(root.children[0].children[0].testResult).To(Equal("FAIL"))
			Expect(root.children[0].children[0].children[0].testResult).To(Equal("FAIL"))
			Expect(root.children[0].children[0].children[1].testResult).To(Equal("PASS"))
		})

		It("calculates correct runtime", func(){
			insertNode(root, []string{"DescribeBlock", "ContextBlock", "SpecBlock1"})
			insertNode(root, []string{"DescribeBlock", "ContextBlock", "SpecBlock2"})
			testSummary1 := &types.SpecSummary{ComponentTexts:[]string{"[Top_Level]", "DescribeBlock", "ContextBlock", "SpecBlock1"}, RunTime:1}
			updateResult(root, testSummary1, "PASS")
			testSummary2 := &types.SpecSummary{ComponentTexts:[]string{"[Top_Level]", "DescribeBlock", "ContextBlock", "SpecBlock2"}, RunTime:1}
			updateResult(root, testSummary2, "PASS")
			Expect(root.children[0].time).To(Equal(time.Duration(2)))
			Expect(root.children[0].children[0].time).To(Equal(time.Duration(2)))
			Expect(root.children[0].children[0].children[0].time).To(Equal(time.Duration(1)))
			Expect(root.children[0].children[0].children[1].time).To(Equal(time.Duration(1)))
		})
	})

	Context("NewAutoGolandReporter", func(){
		It("Returns a goland reporter if OLDPWD contains goland", func(){
			os.Setenv("OLDPWD", "gOlAnD")
			reporter := NewAutoGolandReporter()
			Expect(fmt.Sprintf("%T", reporter)).To(Equal("golandreporter.GolandReporter"))
		})

		It("Returns default reporter if OLDPWD or SNAP_NAME does not contain goland", func(){
			os.Setenv("OLDPWD", "vscode")
			os.Setenv("SNAP_NAME", "vscode")
			reporter := NewAutoGolandReporter()
			Expect(fmt.Sprintf("%T", reporter)).To(Equal("*reporters.DefaultReporter"))
		})

		It("Returns a goland reporter if SNAP_NAME contains goland", func(){
			os.Setenv("SNAP_NAME", "gOlAnD")
			reporter := NewAutoGolandReporter()
			Expect(fmt.Sprintf("%T", reporter)).To(Equal("golandreporter.GolandReporter"))
		})

	})
})
