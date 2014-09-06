package main

import (
	"container/list"
	"fmt"
	"bufio"
	"os"
	"math"
	"strconv"
	"flag"
	"runtime"
	"time"
)

//****************CONST AND STRUCT****************

const (
	TOLERANCE = 0.1
	TOLERANCE_CHECK = 0.2
	BUFFER_SIZE = 20
)

var REPORT_CHANNEL = make(chan bool, 1e6)	
var FINISH_CHANNEL = make(chan bool)

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
	
	if node.isFixed {
		result += ", Fix"
	} else {
		result += ", Non-fix"
	}
	
	for e := node.beams.Front(); e != nil; e = e.Next() {
		beam, _ := e.Value.(*Beam)
		result += "\n\t" + beam.String()
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
	buffer chan float64
}

func (beam *Beam) String() (result string) {
	result = fmt.Sprintf("Beam id: %d\t OtherEndNode: %d df: %.2f cof: %.2f moment: %.1f",
	beam.id, beam.otherEndNode.id, beam.df, beam.cof, beam.moment)
	return result
}

func printStructure(structure *Structure) {
	for _, node := range structure.nodeMap {
		fmt.Println(node.String())
	}
}

//**********************MAIN***********************

func main() {
	//Set Number of Cores
	var numCores = flag.Int("n", 4, "number of CPU cores to use")
	flag.Parse()
	runtime.GOMAXPROCS(*numCores)
	
	filename := "Node1e4.txt"
	
	//Sequential Version===========================================
	structure1 := createStructureFromFile(filename)
	
	start := time.Now()
    
	analyseStructureSequential(structure1)
	
    elapsed := time.Since(start)
    fmt.Printf("Sequential version took %s\n", elapsed)
	
	//Parallel Version=========================================
	structure2 := createStructureFromFile(filename)
	
	start = time.Now()
	
	analyseStructureSynchronous(structure2)
	
    elapsed = time.Since(start)
    fmt.Printf("Parallel version took %s\n", elapsed)
	
	//Check Correctness========================================
	if checkStructure(structure1, structure2) {
		fmt.Println("Same")
	} else {
		fmt.Println("Not Same")
	}
}

//*******************CONSTRUCT STRUCTURE***********

func createStructureFromFile(filename string) (structure *Structure) {
	inputFile, inputError := os.Open(filename)
	if inputError != nil {
		fmt.Println(inputError)
		fmt.Println("An error occurred on opening the inputfile")
		return // exit the function on error
	}
	defer inputFile.Close()
	scanner := bufio.NewScanner(inputFile) //bufio.NewReader(inputFile))
	scanner.Split(bufio.ScanWords)
	
	structure = new(Structure)
	structure.nodeMap = make(map[int]Node)
	
	//read number of nodes
	scanner.Scan()
	numNodes, _ := strconv.Atoi(scanner.Text())
		
	//read nodes
	for i := 0; i < numNodes; i++ {
		scanner.Scan()
		id, _ := strconv.Atoi(scanner.Text())
		scanner.Scan()
		isFixed := scanner.Text() == "F"
		structure.nodeMap[id] = *newNode(id, isFixed)
	}
	
	//read number of beams
	scanner.Scan()
	numBeams, _ := strconv.Atoi(scanner.Text())
		
	//read beams
	for i := 0; i < numBeams; i++ {
		scanner.Scan()
		id1, _ := strconv.Atoi(scanner.Text())
		scanner.Scan()
		df1, _ := strconv.ParseFloat(scanner.Text(), 64)
		scanner.Scan()
		cof1, _ := strconv.ParseFloat(scanner.Text(), 64)
		scanner.Scan()
		moment1, _ := strconv.ParseFloat(scanner.Text(), 64)
		scanner.Scan()
		id2, _ := strconv.Atoi(scanner.Text())
		scanner.Scan()
		df2, _ := strconv.ParseFloat(scanner.Text(), 64)
		scanner.Scan()
		cof2, _ := strconv.ParseFloat(scanner.Text(), 64)
		scanner.Scan()
		moment2, _ := strconv.ParseFloat(scanner.Text(), 64)
		
		connectNodes(structure, id1, df1, cof1, moment1, id2, df2, cof2, moment2)
	}
	
	normalizeStructure(structure)
	
	return
}

func connectNodes(structure *Structure, id1 int, df1 float64, cof1 float64, moment1 float64, id2 int, df2 float64, cof2 float64, moment2 float64) {
	node1 := structure.nodeMap[id1]
	beam1 := new(Beam)
	node2 := structure.nodeMap[id2]
	beam2 := new(Beam)
	
	beam1.df = df1
	beam1.cof = cof1
	beam1.moment = moment1
	beam1.otherEndNode = &node2
	beam1.otherEndBeam = beam2
	beam1.buffer = make(chan float64, BUFFER_SIZE)
	
	beam2.df = df2
	beam2.cof = cof2
	beam2.moment = moment2
	beam2.otherEndNode = &node1
	beam2.otherEndBeam = beam1
	beam2.buffer = make(chan float64, BUFFER_SIZE)
	
	node1.beams.PushBack(beam1)
	node2.beams.PushBack(beam2)
}

func normalizeStructure(structure *Structure) {
	for id, node := range structure.nodeMap { //default order
		if node.beams.Len() > 0 {//!node.isFixed {
			//normalize df
			dfSum := float64(0)
			for e := node.beams.Front(); e != nil; e = e.Next() {
				beam, _ := e.Value.(*Beam)
				dfSum += beam.df
			}
			
			for e := node.beams.Front(); e != nil; e = e.Next() {	
				beam, _ := e.Value.(*Beam)
				beam.df /= dfSum
			}
		} else {
			delete(structure.nodeMap, id)
		}
	}
}

//*******************ANALYZE STRUCTURE SEQUENTIAL**

func analyseStructureSequential(structure *Structure) {
	isFinish := false
	iteration := 0
	for !isFinish {
		iteration++
		isFinish = true
		for _, node := range structure.nodeMap { //default order
			if !node.isFixed {
				//calculate amount of unbalance
				momentSum := float64(0)
				for e := node.beams.Front(); e != nil; e = e.Next() {
					beam, _ := e.Value.(*Beam)
					momentSum += beam.moment
				}

				//redistribute moment and carry over
				if (math.Abs(momentSum) > TOLERANCE) {
					isFinish = false

					for e := node.beams.Front(); e != nil; e = e.Next() {
						beam, _ := e.Value.(*Beam)				
						increment := - momentSum * beam.df
						beam.moment += increment
						beam.otherEndBeam.moment += increment * beam.cof
					}
				}
			}
		}
	}
	fmt.Println("Sequential Analyse Finish, Iteration: ", iteration)
}

//*******************ANALYZE STRUCTURE PARALLEL****

func analyseStructureSynchronous(structure *Structure) {
	//start running
	for id, _ := range structure.nodeMap {
		go analyseNode(structure, id)
	}
	
	All:
	for {
		//time.Sleep(1 * time.Millisecond)
		for count := 0; count < len(structure.nodeMap); count++ {
			select {
			case _, ok := <- REPORT_CHANNEL:
				if ok {
					break
				}
			}
		}
		
		isFinish := true
		Scan:
		for id, _ := range structure.nodeMap { //default order
			for e := structure.nodeMap[id].beams.Front(); e != nil; e = e.Next() {
				beam, _ := e.Value.(*Beam)

				//check pending updated value
				select {
				case value, ok := <-beam.buffer:
					if ok {
						beam.buffer <- value
						isFinish = false
						break Scan
					}
				default:
					break
				}
			}
		}
		
		if isFinish {
			close(FINISH_CHANNEL)
			fmt.Println("Parallel Analyse Finish")
			break All
		}
	}
}

func analyseNode(structure *Structure, id int) {	
	AnalyseNode:
	for {
		//check whether analyse finish
		select {
			case _, ok := <- FINISH_CHANNEL:
				if !ok {
					break AnalyseNode
				}
			default:
				break
		}
		
		if !structure.nodeMap[id].isFixed {
			//calculate amount of unbalance for non-fixed ends
			momentSum := float64(0)
			for e := structure.nodeMap[id].beams.Front(); e != nil; e = e.Next() {
				beam, _ := e.Value.(*Beam)

				//check pending updated value
				select {
				case value, ok := <-beam.buffer:
					if ok {
						beam.moment += value
					} else {
						fmt.Println("Error: Channel closed!")
						break
					}
				default:
					break
				}
				momentSum += beam.moment
			}

			//redistribute moment and carry over
			if (math.Abs(momentSum) > TOLERANCE) {
				for e := structure.nodeMap[id].beams.Front(); e != nil; e = e.Next() {
					beam, _ := e.Value.(*Beam)				
					increment := - momentSum * beam.df
					beam.moment += increment
					beam.otherEndBeam.buffer <- increment * beam.cof
				}
			}
		} else {
			//check updated value for fixed ends
			for e := structure.nodeMap[id].beams.Front(); e != nil; e = e.Next() {
				beam, _ := e.Value.(*Beam)

				select {
				case value, ok := <-beam.buffer:
					if ok {
						beam.moment += value
					} else {
						fmt.Println("Error: Channel closed!")
						break
					}
				default:
					break
				}
			}
		}
		
		REPORT_CHANNEL <- true
		time.Sleep(1 * time.Millisecond)
	}
}

//******************CHECK CORRECTNESS****************

func checkStructure(structure1, structure2 *Structure) (isSame bool) {
	isSame = true
	
	for id, _ := range structure1.nodeMap {
		node1 := structure1.nodeMap[id]
		node2 := structure2.nodeMap[id]
		
		e2 := node2.beams.Front()
		for e1 := node1.beams.Front(); e1 != nil; e1 = e1.Next() {
			moment1 := e1.Value.(*Beam).moment
			moment2 := e2.Value.(*Beam).moment
			if (math.Abs(moment1 - moment2) > TOLERANCE_CHECK) {
				fmt.Println(node1.String())
				fmt.Println(node2.String())
				isSame = false //return
				break
			}
			e2 = e2.Next()
		}
	}
	return isSame
}