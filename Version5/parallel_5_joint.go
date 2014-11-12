package main

import (
	//"container/list"
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
	ends map[int] *End
	buffer chan *Update
}

func (node *Node) String() (result string) {
	result = fmt.Sprintf("Node id: %d, num of ends: %d", node.id, len(node.ends))
	
	if node.isFixed {
		result += ", Fix"
	} else {
		result += ", Non-fix"
	}
	
	for _, end := range node.ends {
		result += "\n\t" + end.String()
	}
	return result
}

func (node *Node) addEnd(end *End) (endIndex int) {
	endIndex = len(node.ends);
	node.ends[endIndex] = end;
	return endIndex;
} 

func newNode(id int, isFixed bool) *Node {
	node := new(Node)
	
	node.id = id
	node.isFixed = isFixed
	node.ends = make(map[int]*End)
	node.buffer = make(chan *Update, BUFFER_SIZE)
	return node
}

type Update struct {
	carryover float64
	endIndex int
}

func newUpdate(endIndex int, carryover float64) *Update {
	update := new(Update)
	
	update.endIndex = endIndex
	update.carryover = carryover
	
	return update
}

type End struct {
	otherEndNodeID int
	otherEndIndex int
	df float64
	moment float64
}

func (end *End) String() (result string) {
	result = fmt.Sprintf("End df: %.2f moment: %.1f", end.df, end.moment)
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

	analyseStructureAsynchronous(structure2)

	elapsed = time.Since(start)
	fmt.Printf("Parallel version took %s\n", elapsed)
	
	//printStructure(structure2)
	
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
	scanner := bufio.NewScanner(inputFile) 
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
	
	//read number of ends
	scanner.Scan()
	numEnds, _ := strconv.Atoi(scanner.Text())
		
	//read ends
	for i := 0; i < numEnds; i++ {
		scanner.Scan()
		id1, _ := strconv.Atoi(scanner.Text())
		scanner.Scan()
		df1, _ := strconv.ParseFloat(scanner.Text(), 64)
		scanner.Scan()
		//_ := strconv.ParseFloat(scanner.Text(), 64)
		scanner.Scan()
		moment1, _ := strconv.ParseFloat(scanner.Text(), 64)
		scanner.Scan()
		id2, _ := strconv.Atoi(scanner.Text())
		scanner.Scan()
		df2, _ := strconv.ParseFloat(scanner.Text(), 64)
		scanner.Scan()
		//_ := strconv.ParseFloat(scanner.Text(), 64)
		scanner.Scan()
		moment2, _ := strconv.ParseFloat(scanner.Text(), 64)
		
		connectNodes(structure, id1, df1, moment1, id2, df2, moment2)
	}
	
	normalizeStructure(structure)
	
	return
}

func connectNodes(structure *Structure, id1 int, df1 float64, moment1 float64, id2 int, df2 float64, moment2 float64) {
	node1 := structure.nodeMap[id1]
	end1 := new(End)
	
	node2 := structure.nodeMap[id2]
	end2 := new(End)
	
	end1.df = df1
	end1.moment = moment1
	end1.otherEndNodeID = id2
	end1.otherEndIndex = node2.addEnd(end2)
	
	end2.df = df2
	end2.moment = moment2
	end2.otherEndNodeID = id1
	end2.otherEndIndex = node1.addEnd(end1)
}

func normalizeStructure(structure *Structure) {
	for id, node := range structure.nodeMap { //default order
		if len(node.ends) > 0 {
			//normalize df
			dfSum := float64(0)
			for _, end := range node.ends {
				dfSum += end.df
			}
			
			for _, end := range node.ends {
				end.df /= dfSum
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
			moreUpdate := true
			for moreUpdate {
				//check pending updated value
					select {
					case update, ok := <-node.buffer:
						if ok {
							node.ends[update.endIndex].moment += update.carryover
						} else {
							fmt.Println("Error: Channel closed!")
							break
						}
					default:
						moreUpdate = false
						break
				}
			}
			
			if !node.isFixed {
				//calculate amount of unbalance
				momentSum := float64(0)
				for _, end := range node.ends {
					momentSum += end.moment
				}

				//redistribute moment and carry over
				if (math.Abs(momentSum) > TOLERANCE) {
					isFinish = false

					for _, end := range node.ends {			
						increment := - momentSum * end.df
						end.moment += increment
						structure.nodeMap[end.otherEndNodeID].buffer <- newUpdate(end.otherEndIndex, increment * 0.5)
					}
				}
			}
		}
	}
	fmt.Println("Sequential Analyse Finish, Iteration: ", iteration)
}

//*******************ANALYZE STRUCTURE PARALLEL****

func analyseStructureAsynchronous(structure *Structure) {
	idSets := make([]map[int]bool, 4)
	idSets[0] = make(map[int]bool)
	idSets[1] = make(map[int]bool)
	idSets[2] = make(map[int]bool)
	idSets[3] = make(map[int]bool)
	rotate := 0

	//assign nodes to threads
	for id, _ := range structure.nodeMap {
		idSets[rotate][id] = true
		rotate = (rotate + 1) % 4
	}

	//start parallel analysis
	for i := 0; i < 4; i++ {
		go analyseNode(structure, idSets[i])
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
		for _, node := range structure.nodeMap { //default order
			//check pending updated value
			select {
			case value, ok := <-node.buffer:
				if ok {
					node.buffer <- value
					isFinish = false
					break Scan
				}
			default:
				break
			}
		}

		if isFinish {
			close(FINISH_CHANNEL)
			fmt.Println("Parallel Analyse Finish")
			break All
		}
	}
}

func analyseNode(structure *Structure, idSet map[int]bool) {
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

		for id, _ := range idSet {
			node := structure.nodeMap[id]

			moreUpdate := true
			for moreUpdate {
				//check pending updated value
					select {
					case update, ok := <-node.buffer:
						if ok {
							node.ends[update.endIndex].moment += update.carryover
						} else {
							fmt.Println("Error: Channel closed!")
							break
						}
					default:
						moreUpdate = false
						break
				}
			}

			if !node.isFixed {
				//calculate amount of unbalance for non-fixed ends
				momentSum := float64(0)
				for _, end := range structure.nodeMap[id].ends {
					momentSum += end.moment
				}

				//redistribute moment and carry over
				if (math.Abs(momentSum) > TOLERANCE) {
					for _, end := range structure.nodeMap[id].ends {
						increment := - momentSum * end.df
						end.moment += increment
						structure.nodeMap[end.otherEndNodeID].buffer <- newUpdate(end.otherEndIndex, increment * 0.5)
					}
				}
			}
		}
		REPORT_CHANNEL <- true
		//time.Sleep(1 * time.Millisecond)
	}
}

//******************CHECK CORRECTNESS****************

func checkStructure(structure1, structure2 *Structure) (isSame bool) {
	isSame = true
	
	for id, _ := range structure1.nodeMap {
		node1 := structure1.nodeMap[id]
		node2 := structure2.nodeMap[id]
		
		for endIndex, _ := range node1.ends {
			moment1 := node1.ends[endIndex].moment
			moment2 := node2.ends[endIndex].moment
			if (math.Abs(moment1 - moment2) > TOLERANCE_CHECK) {
				fmt.Println(node1.String())
				fmt.Println(node2.String())
				isSame = false //return
				break
			}
		}
	}
	return isSame
}