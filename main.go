package main

import (
	"fmt"
	"sync"
)

func printOdd(n int, oddChan, evenChan chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 1; i <= n; i += 2 {
		<-oddChan // 1. Wait for our turn
		fmt.Printf("Odd : %d\n", i)
		if i+1 <= n {
			evenChan <- struct{}{} // 2. Signal the even goroutine
		}
	}
}

func printEven(n int, oddChan, evenChan chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := 2; i <= n; i += 2 {
		<-evenChan // 1. Wait for our turn
		fmt.Printf("Even : %d\n", i)
		if i+1 <= n {
			oddChan <- struct{}{} // 2. Signal the odd goroutine
		}
	}
}

func main() {
	n := 10
	oddChan := make(chan struct{})
	evenChan := make(chan struct{})
	var wg sync.WaitGroup

	fmt.Println("Starting of the program")

	wg.Add(2)
	go printOdd(n, oddChan, evenChan, &wg)
	go printEven(n, oddChan, evenChan, &wg)

	// Kick off the sequence by signaling the odd goroutine first
	oddChan <- struct{}{}

	wg.Wait()
	fmt.Println("End of the program")
}
