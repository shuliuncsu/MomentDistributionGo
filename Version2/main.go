package main

import (
	//"errors"
	"fmt"
	"math"
)

//****************CONST AND STRUCT****************

const (
	MAX_NUM_BEAMS = 2
	NUM_NODES = 3
	NUM_BEAMS = 2
	TOLERANCE = 0.1
)

type structNode struct {
	id int
	isFixed bool
	numBeams int
	beams []*structBeam
}

func (node *structNode) String() (result string) {
	result = fmt.Sprintf("Node id: %d, num of beams: %d", node.id, node.numBeams)
	for i := 0; i < node.numBeams; i++ {
		result += "\n\t" + node.beams[i].String()
	}
	return result
}

func newNode() *structNode {
	node := new(structNode)
	
	node.id = 0
	node.isFixed = false
	node.numBeams = 0
	node.beams = new([MAX_NUM_BEAMS]*structBeam)[:]
	for i := 0; i < MAX_NUM_BEAMS; i++ {
		node.beams[i] = new(structBeam)
	}
	return node
}

type structBeam struct {
	id int
	otherEndNode *structNode
	otherEndBeam *structBeam
	df float64
	cof float64 //cof to the other end
	moment float64
}

func (beam *structBeam) String() (result string) {
	result = fmt.Sprintf("Beam id: %d\t OtherEndNode: %d df: %.2f cof: %.2f moment: %.1f",
		beam.id, beam.otherEndNode.id, beam.df, beam.cof, beam.moment)
	return result
}

//**********************MAIN**********************

func main() {
	nodes := createStucture()
	
	configStructureExample1(nodes)
	
	printStructure(nodes)
	
	analyseStructure(nodes)
}

//**********************PRINT*********************

func printStructure(nodes []structNode) {
	printNodes(nodes)
}

func printNodes(nodes []structNode) {
	fmt.Println("<Nodes>")
	for i := 0; i < len(nodes); i++ {
		fmt.Println(nodes[i].String())
	}
	fmt.Println()
}

//*******************CONSTRUCT STRUCTURE***********

func createStucture() (nodes []structNode) {
	//create nodes
	nodes = make([]structNode, NUM_NODES)
	for i := 0; i < NUM_NODES; i++ {
		nodes[i] = *newNode()
		nodes[i].id = i
	}

	//connect nodes linearly
	for i := 0; i < NUM_NODES - 1; i++ {
		connectNodes(&nodes[i], &nodes[i + 1])
	}
	return
}

func connectNodes(node1 *structNode, node2 *structNode) {

	node1.beams[node1.numBeams].otherEndNode = node2
	node1.beams[node1.numBeams].id = node1.numBeams + 1

	node2.beams[node2.numBeams].otherEndNode = node1
	node2.beams[node2.numBeams].id = node2.numBeams + 1

	node1.beams[node1.numBeams].otherEndBeam = node2.beams[node2.numBeams]
	node2.beams[node2.numBeams].otherEndBeam = node1.beams[node1.numBeams]

	node1.numBeams++
	node2.numBeams++
}

func configStructureExample1(nodes []structNode) {
	nodes[0].isFixed = true
	
	//beams[0].df = 0
	//beams[0].cof = 0
	nodes[0].beams[0].moment = -172.8
	
	nodes[1].beams[0].df = 0.5
	nodes[1].beams[0].cof = 0.5
	nodes[1].beams[0].moment = 115.2

	nodes[1].beams[1].df = 0.5
	nodes[1].beams[1].cof = 0.5
	nodes[1].beams[1].moment = -416.7
	
	nodes[2].beams[0].df = 1.0
	nodes[2].beams[0].cof = 0.5
	nodes[2].beams[0].moment = 416.7
}

//*******************ANALYZE STRUCTURE*************

func analyseStructure(nodes []structNode) {
	isFinish := false
	iteration := 0
	for !isFinish {
		iteration++
		isFinish = true
		for nodeIndex := 0; nodeIndex < len(nodes); nodeIndex++ { //normal order
		//for nodeIndex := len(nodes) - 1; nodeIndex >= 0; nodeIndex-- { //reverse order
			if !nodes[nodeIndex].isFixed {
				//calculate amount of unbalance
				momentSum := float64(0)
				for beamIndex := 0; beamIndex < nodes[nodeIndex].numBeams; beamIndex++ {
					momentSum += nodes[nodeIndex].beams[beamIndex].moment
				}

				//redistribute moment and carry over
				if (math.Abs(momentSum) > TOLERANCE) {
					isFinish = false
					
					for beamIndex := 0; beamIndex < nodes[nodeIndex].numBeams; beamIndex++ {
						increment := - momentSum * nodes[nodeIndex].beams[beamIndex].df
						nodes[nodeIndex].beams[beamIndex].moment += increment
							
						nodes[nodeIndex].beams[beamIndex].otherEndBeam.moment +=
							increment * nodes[nodeIndex].beams[beamIndex].cof
					}
				}
			}
		}
		fmt.Println("\nIteration: ", iteration)
		printStructure(nodes)
	}
}