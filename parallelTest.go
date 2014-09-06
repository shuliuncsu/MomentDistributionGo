package main

import (
	"fmt"
	"flag"
	"runtime"
	"time"
)

const (
	MAX = 1000000
)

func main() {
	//Set Number of Cores
	var numCores = flag.Int("n", 4, "number of CPU cores to use")
	flag.Parse()
	runtime.GOMAXPROCS(*numCores)
	
	//Sequential Version===========================================	
	start := time.Now()
    
	sumSequential()
	
    elapsed := time.Since(start)
    fmt.Printf("Sequential version took %s\n", elapsed)
	
	//Parallel Version=========================================	
	start = time.Now()
	
	sumParallel()
	
    elapsed = time.Since(start)
    fmt.Printf("Parallel version took %s\n", elapsed)
}

func sumSequential() {
	total := 0
	for i := 1; i <= MAX; i++ {
		total += i
	}
	fmt.Println(total)
}

func sumParallel() {
	total := 0
	sum := make(chan int, 4)
	
	go sumHelper(1, 				MAX / 4, 		sum)
	go sumHelper(MAX / 4 + 1, 		MAX / 2, 		sum)
	go sumHelper(MAX / 2 + 1, 		MAX * 3 / 4, 	sum)
	go sumHelper(MAX * 3 / 4 + 1, 	MAX, 			sum)
	
	for i := 1; i <= 4; i++ {
		select {
		case value, _ := <-sum:
			total += value
		}
	}
	
	fmt.Println(total)
}

func sumHelper(start, end int, sum chan int) {
	result := 0
	for i := start; i <= end; i++ {
		result += i
	}
	sum <- result
}