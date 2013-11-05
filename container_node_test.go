package ginkgo

import (
	"github.com/onsi/ginkgo/types"
	. "github.com/onsi/gomega"
	"math/rand"
	"sort"
	"time"
)

func init() {
	Describe("Container Node", func() {
		var (
			codeLocation types.CodeLocation
			container    *containerNode
		)

		BeforeEach(func() {
			codeLocation = types.GenerateCodeLocation(0)
			container = newContainerNode("description text", flagTypeFocused, codeLocation)
		})

		Describe("creating a container node", func() {
			It("stores off the passed in properties", func() {
				Ω(container.text).Should(Equal("description text"))
				Ω(container.flag).Should(Equal(flagTypeFocused))
				Ω(container.codeLocation).Should(Equal(codeLocation))
			})
		})

		Describe("appending", func() {
			Describe("it nodes", func() {
				Context("appending nodes", func() {
					var itA, itB *itNode
					var subContainer *containerNode
					BeforeEach(func() {
						itA = newItNode("itA", func() {}, flagTypeNone, types.GenerateCodeLocation(0), 0)
						itB = newItNode("itB", func() {}, flagTypeNone, types.GenerateCodeLocation(0), 0)
						subContainer = newContainerNode("subcontainer", flagTypeNone, types.GenerateCodeLocation(0))
						container.pushSubjectNode(itA)
						container.pushContainerNode(subContainer)
						container.pushSubjectNode(itB)
					})
					It("can append container nodes and it nodes", func() {
						Ω(container.subjectAndContainerNodes).Should(Equal([]node{
							itA,
							subContainer,
							itB,
						}))
					})
					It("will append only subject nodes to subjects list", func() {
						Ω(container.subjects).Should(Equal([]exampleSubject{
							itA,
							itB,
						}))
					})
					It("will signal container that subjects have run", func() {
						bex := newExample(itB)
						bex.addContainerNode(container)
						bex.runSample(1)
						Ω(container.subjectRunCount).Should(BeNumerically("==", 1))
					})
					Describe("run once containers", func() {
						var bex, aex *example
						BeforeEach(func() {
							bex = newExample(itB)
							bex.addContainerNode(container)
							aex = newExample(itA)
							aex.addContainerNode(container)
						})
						Context("BeforeAll is specified", func() {
							var beforeAllRun = false
							BeforeEach(func() {
								beforeAllRun = false
								container.pushBeforeAllNode(
									newRunnableNode(func() {
										beforeAllRun = true
									},
									types.GenerateCodeLocation(1),
									6*time.Second),
								)
							})
							It("beforeAll runs when first contained element is run", func(){
								bex.runSample(1)
								Ω(beforeAllRun).Should(BeTrue())
							})
							It("beforeAll runs only 1 time even if 2 subjects run", func(){
								bex.runSample(1)
								Ω(beforeAllRun).Should(BeTrue())
								beforeAllRun = false
								aex.runSample(1)
								Ω(beforeAllRun).Should(BeFalse())
							})
						})
						Context("AfterAll is specified", func() {
							var afterAllRun = false
							BeforeEach(func() {
								afterAllRun = false
								container.pushAfterAllNode(
									newRunnableNode(func() {
										afterAllRun = true
									},
										types.GenerateCodeLocation(1),
										6*time.Second),
								)
							})
							It("afterAll runs if both containers run", func() {
								bex.runSample(1)
								Ω(afterAllRun).Should(BeFalse())
								aex.runSample(1)
								Ω(afterAllRun).Should(BeTrue())
							})
							It("afterAll runs if subject panics", func() {
								panicking := &panickySubject{}
								container.pushSubjectNode(panicking)
								bex.runSample(1)
								Ω(afterAllRun).Should(BeFalse())
								aex.runSample(1)
								pex := newExample(panicking)
								pex.addContainerNode(container)
								defer func() {
									if r := recover(); r != nil {
										if !Ω(afterAllRun).Should(BeTrue()) {
											panic(r)
										}
									}
								}()
								pex.runSample(1)
							})
						})
					})
				})
			})

			Describe("other runnable nodes", func() {
				var (
					runnableA *runnableNode
					runnableB *runnableNode
				)

				BeforeEach(func() {
					runnableA = newRunnableNode(func() {}, types.GenerateCodeLocation(0), 0)
					runnableB = newRunnableNode(func() {}, types.GenerateCodeLocation(0), 0)
				})

				It("can append multiple beforeEach nodes", func() {
					container.pushBeforeEachNode(runnableA)
					container.pushBeforeEachNode(runnableB)
					Ω(container.beforeEachNodes).Should(Equal([]*runnableNode{
						runnableA,
						runnableB,
					}))
				})

				It("can append multiple justBeforeEach nodes", func() {
					container.pushJustBeforeEachNode(runnableA)
					container.pushJustBeforeEachNode(runnableB)
					Ω(container.justBeforeEachNodes).Should(Equal([]*runnableNode{
						runnableA,
						runnableB,
					}))
				})

				It("can append multiple afterEach nodes", func() {
					container.pushAfterEachNode(runnableA)
					container.pushAfterEachNode(runnableB)
					Ω(container.afterEachNodes).Should(Equal([]*runnableNode{
						runnableA,
						runnableB,
					}))
				})
			})
		})

		Describe("generating examples", func() {
			var (
				itA          *itNode
				itB          *itNode
				subContainer *containerNode
				subItA       *itNode
				subItB       *itNode
			)

			BeforeEach(func() {
				itA = newItNode("itA", func() {}, flagTypeNone, types.GenerateCodeLocation(0), 0)
				itB = newItNode("itB", func() {}, flagTypeNone, types.GenerateCodeLocation(0), 0)
				subContainer = newContainerNode("subcontainer", flagTypeNone, types.GenerateCodeLocation(0))
				subItA = newItNode("subItA", func() {}, flagTypeNone, types.GenerateCodeLocation(0), 0)
				subItB = newItNode("subItB", func() {}, flagTypeNone, types.GenerateCodeLocation(0), 0)

				container.pushSubjectNode(itA)
				container.pushContainerNode(subContainer)
				container.pushSubjectNode(itB)

				subContainer.pushSubjectNode(subItA)
				subContainer.pushSubjectNode(subItB)
			})

			It("generates an example for each It in the hierarchy", func() {
				examples := container.generateExamples()
				Ω(examples).Should(HaveLen(4))

				Ω(examples[0].subject).Should(Equal(itA))
				Ω(examples[0].containers).Should(Equal([]*containerNode{container}))

				Ω(examples[1].subject).Should(Equal(subItA))
				Ω(examples[1].containers).Should(Equal([]*containerNode{container, subContainer}))

				Ω(examples[2].subject).Should(Equal(subItB))
				Ω(examples[2].containers).Should(Equal([]*containerNode{container, subContainer}))

				Ω(examples[3].subject).Should(Equal(itB))
				Ω(examples[3].containers).Should(Equal([]*containerNode{container}))
			})

			It("ignores containers in the hierarchy that are empty", func() {
				emptyContainer := newContainerNode("empty container", flagTypeNone, types.GenerateCodeLocation(0))
				emptyContainer.pushBeforeEachNode(newRunnableNode(func() {}, types.GenerateCodeLocation(0), 0))

				container.pushContainerNode(emptyContainer)
				examples := container.generateExamples()
				Ω(examples).Should(HaveLen(4))
			})
		})

		Describe("shuffling the container", func() {
			texts := func(container *containerNode) []string {
				texts := make([]string, 0)
				for _, node := range container.subjectAndContainerNodes {
					texts = append(texts, node.getText())
				}
				return texts
			}

			BeforeEach(func() {
				itA := newItNode("Banana", func() {}, flagTypeNone, types.GenerateCodeLocation(0), 0)
				itB := newItNode("Apple", func() {}, flagTypeNone, types.GenerateCodeLocation(0), 0)
				itC := newItNode("Orange", func() {}, flagTypeNone, types.GenerateCodeLocation(0), 0)
				containerA := newContainerNode("Cucumber", flagTypeNone, types.GenerateCodeLocation(0))
				containerB := newContainerNode("Airplane", flagTypeNone, types.GenerateCodeLocation(0))

				container.pushSubjectNode(itA)
				container.pushContainerNode(containerA)
				container.pushSubjectNode(itB)
				container.pushContainerNode(containerB)
				container.pushSubjectNode(itC)
			})

			It("should be sortable", func() {
				sort.Sort(container)
				Ω(texts(container)).Should(Equal([]string{"Airplane", "Apple", "Banana", "Cucumber", "Orange"}))
			})

			It("shuffles all the examples after sorting them", func() {
				container.shuffle(rand.New(rand.NewSource(17)))
				expectedOrder := shuffleStrings([]string{"Airplane", "Apple", "Banana", "Cucumber", "Orange"}, 17)
				Ω(texts(container)).Should(Equal(expectedOrder), "The permutation should be the same across test runs")
			})
		})
	})
}

type panickySubject struct {
}

func (node *panickySubject) generateExamples() []*example {
	return []*example{newExample(node)}
}
func (node *panickySubject) nodeType() nodeType {
	return nodeTypeIt
}
func (node *panickySubject) getText() string {
	return "some text?"
}
func (node *panickySubject) getFlag() flagType {
	return flagTypeNone
}
func (node *panickySubject) getCodeLocation() types.CodeLocation {
	return types.GenerateCodeLocation(1)
}
func (node *panickySubject) run() (outcome runOutcome, failure failureData) {
	panic("i'm panicky.  so I panic!")
}
