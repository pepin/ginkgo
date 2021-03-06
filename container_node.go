package ginkgo

import (
	"github.com/onsi/ginkgo/types"
	"math/rand"
	"sort"
)

type containerNode struct {
	flag         flagType
	text         string
	codeLocation types.CodeLocation
	subjectRunCount int

	beforeAllNodes           []*runnableNode
	beforeEachNodes          []*runnableNode
	justBeforeEachNodes      []*runnableNode
	afterEachNodes           []*runnableNode
	afterAllNodes            []*runnableNode
	subjectAndContainerNodes []node
	subjects                 []exampleSubject
}

func newContainerNode(text string, flag flagType, codeLocation types.CodeLocation) *containerNode {
	return &containerNode{
		text:         text,
		flag:         flag,
		codeLocation: codeLocation,
	}
}

func (container *containerNode) shuffle(r *rand.Rand) {
	sort.Sort(container)
	permutation := r.Perm(len(container.subjectAndContainerNodes))
	shuffledNodes := make([]node, len(container.subjectAndContainerNodes))
	for i, j := range permutation {
		shuffledNodes[i] = container.subjectAndContainerNodes[j]
	}
	container.subjectAndContainerNodes = shuffledNodes
}

func (node *containerNode) generateExamples() []*example {
	examples := make([]*example, 0)

	for _, containerOrSubject := range node.subjectAndContainerNodes {
		examples = append(examples, containerOrSubject.generateExamples()...)
	}

	for _, example := range examples {
		example.addContainerNode(node)
	}

	return examples
}

func (node *containerNode) AllSubjectsRun() bool {
	return len(node.subjects) == node.subjectRunCount
}
func (node *containerNode) NoSubjectsRun() bool {
	return 0 == node.subjectRunCount
}
func (node *containerNode) NotifyComplete(subject exampleSubject) {
	for _, s := range node.subjects {
		if s == subject {
			node.subjectRunCount += 1
			return
		}
	}
}

func (node *containerNode) pushContainerNode(container *containerNode) {
	node.subjectAndContainerNodes = append(node.subjectAndContainerNodes, container)
}

func (node *containerNode) pushSubjectNode(subject exampleSubject) {
	node.subjectAndContainerNodes = append(node.subjectAndContainerNodes, subject)
	node.subjects = append(node.subjects, subject)
}

func (node *containerNode) pushAfterAllNode(afterAll *runnableNode) {
	node.afterAllNodes = append(node.afterAllNodes, afterAll)
}

func (node *containerNode) pushBeforeAllNode(beforeAll *runnableNode) {
	node.beforeAllNodes = append(node.beforeAllNodes, beforeAll)
}

func (node *containerNode) pushBeforeEachNode(beforeEach *runnableNode) {
	node.beforeEachNodes = append(node.beforeEachNodes, beforeEach)
}

func (node *containerNode) pushJustBeforeEachNode(justBeforeEach *runnableNode) {
	node.justBeforeEachNodes = append(node.justBeforeEachNodes, justBeforeEach)
}

func (node *containerNode) pushAfterEachNode(afterEach *runnableNode) {
	node.afterEachNodes = append(node.afterEachNodes, afterEach)
}

func (node *containerNode) nodeType() nodeType {
	return nodeTypeContainer
}

func (node *containerNode) getText() string {
	return node.text
}

//sort.Interface

func (node *containerNode) Len() int {
	return len(node.subjectAndContainerNodes)
}

func (node *containerNode) Less(i, j int) bool {
	return node.subjectAndContainerNodes[i].getText() < node.subjectAndContainerNodes[j].getText()
}

func (node *containerNode) Swap(i, j int) {
	node.subjectAndContainerNodes[i], node.subjectAndContainerNodes[j] = node.subjectAndContainerNodes[j], node.subjectAndContainerNodes[i]
}
