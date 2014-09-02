package main

import (
	"container/list"
	"fmt"
	//"math"
)

//****************CONST AND STRUCT****************

const (
	TOLERANCE = 0.001
)

type Structure struct {
	nodeMap map[int]Node
}


type Node struct {
	id int
	isFixed bool
	beams *list.List
}

func (node *Node) String() (result string) {
	result = fmt.Sprintf("Node id: %d, num of beams: %d", node.id, node.beams.Len())
	
	for e := node.beams.Front(); e != nil; e = e.Next() {
		if beam, ok := e.Value.(Beam); ok {
			result += "\n\t" + beam.String()
		}
	}
	return result
}

func newNode(id int, isFixed bool) *Node {
	node := new(Node)
	
	node.id = id
	node.isFixed = isFixed
	node.beams = list.New()
	return node
}

type Beam struct {
	id int
	otherEndNode *Node
	otherEndBeam *Beam
	df float64
	cof float64 //cof to the other end
	moment float64
}

func (beam *Beam) String() (result string) {
	result = fmt.Sprintf("Beam id: %d\t OtherEndNode: %d df: %.2f cof: %.2f moment: %.1f",
		beam.id, beam.otherEndNode.id, beam.df, beam.cof, beam.moment)
	return result
}

//**********************MAIN**********************

func main() {
	structure := createStucture()
		
	printStructure(structure)
	
	analyseStructure(structure)
}

//**********************PRINT*********************

func printStructure(structure *Structure) {
	for _, node := range structure.nodeMap {
		fmt.Println(node.String())
	}
}

//*******************CONSTRUCT STRUCTURE***********

func createStucture() (structure *Structure) {
	structure = new(Structure)
	return
}

//*******************ANALYZE STRUCTURE*************

func analyseStructure(structure *Structure) {
}

// func analyseStructure(structure *Structure) {
// 	isFinish := false
// 	iteration := 0
// 	for !isFinish {
// 		iteration++
// 		isFinish = true
// 		for nodeIndex := 0; nodeIndex < len(nodes); nodeIndex++ { //normal order
// 		//for nodeIndex := len(nodes) - 1; nodeIndex >= 0; nodeIndex-- { //reverse order
// 			if !nodes[nodeIndex].isFixed {
// 				//calculate amount of unbalance
// 				momentSum := float64(0)
// 				for beamIndex := 0; beamIndex < nodes[nodeIndex].numBeams; beamIndex++ {
// 					momentSum += nodes[nodeIndex].beams[beamIndex].moment
// 				}
//
// 				//redistribute moment and carry over
// 				if (math.Abs(momentSum) > TOLERANCE) {
// 					isFinish = false
//
// 					for beamIndex := 0; beamIndex < nodes[nodeIndex].numBeams; beamIndex++ {
// 						increment := - momentSum * nodes[nodeIndex].beams[beamIndex].df
// 						nodes[nodeIndex].beams[beamIndex].moment += increment
//
// 						nodes[nodeIndex].beams[beamIndex].otherEndBeam.moment +=
// 							increment * nodes[nodeIndex].beams[beamIndex].cof
// 					}
// 				}
// 			}
// 		}
// 		fmt.Println("\nIteration: ", iteration)
// 		printStructure(nodes)
// 	}
// }