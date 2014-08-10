package main

import (
	"errors"
	"fmt"
	"math"
)

const (
	MAX_NUM_BEAMS = 2
	NUM_NODES = 3
	NUM_BEAMS = 2
	END_SIDE_1 = 0
	END_SIDE_2 = 1
	OTHER_SIDE = 1
	TOLERANCE = 0.1
)

func main() {
	nodes, beams := createStucture()
	printStructure(nodes, beams)
	configStructureExample1(nodes, beams)
	printStructure(nodes, beams)
	analyseStructureJacobi(nodes, beams)
}

type structNode struct {
	id int
	isFixed bool
	numBeams int
	beams []*structBeam
	sides []int
}

func (node *structNode) String() (result string) {
	result = fmt.Sprintf("Node id: %d, num of beams: %d", node.id, node.numBeams)
	return result
}

func (node *structNode) addBeam(beam *structBeam, endSide int) (err error) {
	if (node.numBeams < MAX_NUM_BEAMS) {
		node.beams[node.numBeams] = beam
		node.sides[node.numBeams] = endSide
		node.numBeams++
		
		beam.node[endSide] = node
		
		return nil
	} else {
		return errors.New("Node is full")
	}
}

func newNode() *structNode {
	
	node := new(structNode)
	
	node.id = 0
	node.isFixed = false
	node.numBeams = 0
	node.beams = new([MAX_NUM_BEAMS]*structBeam)[:]
	node.sides = new([MAX_NUM_BEAMS]int)[:]
	
	return node
}

type structBeam struct {
	id int
	node [2]*structNode
	df [2]float64
	cof [2]float64
	moment [2]float64
}

func (beam *structBeam) String() (result string) {
	result = fmt.Sprintf("Beam id: %d\nnode[END_SIDE_1]: %d df[END_SIDE_1]: %.2f cof[END_SIDE_1]: %.2f moment[END_SIDE_1]: %.1f\nnode[END_SIDE_2]: %d df[END_SIDE_2]: %4.2f cof[END_SIDE_2]: %4.2f moment[END_SIDE_2]: %.1f",
	 beam.id, beam.node[END_SIDE_1].id, beam.df[END_SIDE_1], beam.cof[END_SIDE_1], beam.moment[END_SIDE_1], beam.node[END_SIDE_2].id, beam.df[END_SIDE_2], beam.cof[END_SIDE_2], beam.moment[END_SIDE_2])
	return result
}

func createStucture() (nodes []structNode, beams []structBeam) {
	nodes = make([]structNode, NUM_NODES)
	for i := 0; i < NUM_NODES; i++ {
		nodes[i] = *newNode()
		nodes[i].id = i
	}
	
	beams = make([]structBeam, NUM_BEAMS)
	for i := 0; i < NUM_BEAMS; i++ {
		//beams[i] = newBeam()
		beams[i].id = i
		nodes[i].addBeam(&beams[i], END_SIDE_1)
		nodes[i + 1].addBeam(&beams[i], END_SIDE_2)
	}
	
	for i := 0; i < NUM_NODES; i++ {
		fmt.Println(i, " ", nodes[i].numBeams)
	}
	
	return
}

func configStructureExample1(nodes []structNode, beams []structBeam) {
	nodes[0].isFixed = true
	
	//beams[0].df[END_SIDE_1] = 0
	//beams[0].cof[END_SIDE_1] = 0
	beams[0].moment[END_SIDE_1] = -172.8
	
	beams[0].df[END_SIDE_2] = 0.5
	beams[0].cof[END_SIDE_2] = 0.5
	beams[0].moment[END_SIDE_2] = 115.2

	beams[1].df[END_SIDE_1] = 0.5
	beams[1].cof[END_SIDE_1] = 0.5
	beams[1].moment[END_SIDE_1] = -416.7
	
	beams[1].df[END_SIDE_2] = 1.0
	beams[1].cof[END_SIDE_2] = 0.5
	beams[1].moment[END_SIDE_2] = 416.7
}

func analyseStructureJacobi(nodes []structNode, beams []structBeam) {
	isFinish := false
	iteration := 0
	for !isFinish {
	//for iteration := 0; iteration < 3; iteration++ {
		iteration++
		isFinish = true
		//for nodeIndex := 0; nodeIndex < len(nodes); nodeIndex++ {
		for nodeIndex := len(nodes) - 1; nodeIndex >= 0; nodeIndex-- {
			if !nodes[nodeIndex].isFixed {
				//calculate amount of unbalance
				momentSum := float64(0)
				for beamIndex := 0; beamIndex < nodes[nodeIndex].numBeams; beamIndex++ {
					momentSum += nodes[nodeIndex].beams[beamIndex].moment[nodes[nodeIndex].sides[beamIndex]]
				}

				//redistribute moment and carry over
				if (math.Abs(momentSum) > TOLERANCE) {
					isFinish = false
					
					for beamIndex := 0; beamIndex < nodes[nodeIndex].numBeams; beamIndex++ {
						increment := - momentSum * nodes[nodeIndex].beams[beamIndex].df[nodes[nodeIndex].sides[beamIndex]]
						nodes[nodeIndex].beams[beamIndex].moment[nodes[nodeIndex].sides[beamIndex]] += increment
							
						nodes[nodeIndex].beams[beamIndex].moment[OTHER_SIDE - nodes[nodeIndex].sides[beamIndex]] +=
							increment * nodes[nodeIndex].beams[beamIndex].cof[nodes[nodeIndex].sides[beamIndex]]
					}
					fmt.Println("\nIteration: ", iteration)
					printBeams(beams)
				}
			}
		}
	}
}

func printStructure(nodes []structNode, beams []structBeam) {
	printNodes(nodes)
	printBeams(beams)
}

func printNodes(nodes []structNode) {
	fmt.Println("<Nodes>")
	for i := 0; i < len(nodes); i++ {
		fmt.Println(nodes[i].String())
	}
	fmt.Println()
}

func printBeams(beams []structBeam) {
	fmt.Println("<Beams>")
	for i := 0; i < len(beams); i++ {
		fmt.Println(beams[i].String())
	}
	fmt.Println()
}
