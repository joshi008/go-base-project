package main

import (
	"fmt"
	"strconv"
)

func fizzF(c chan string) {
	c <- "Fizz"
}

func buzzF(c chan string) {
	c <- "Buzz"
}

func fizzBuzzF(c chan string) {
	c <- "FizzBuzz"
}

func numF(c chan string, num int) {
	// Convert the integer to a string to match the channel type
	c <- strconv.Itoa(num)
}

func main() {
	num := 20

	fizz := make(chan string)
	buzz := make(chan string)
	fizzBuzz := make(chan string)
	numChan := make(chan string)

	for i := 1; i <= num; i++ {
		// 1. Must check 15 (3 and 5) first!
		if i%15 == 0 {
			go fizzBuzzF(fizzBuzz)
			fmt.Println(<-fizzBuzz)

		} else if i%3 == 0 {
			go fizzF(fizz)
			fmt.Println(<-fizz)

		} else if i%5 == 0 {
			go buzzF(buzz)
			fmt.Println(<-buzz)

		} else {
			go numF(numChan, i)
			fmt.Println(<-numChan)
		}
	}
}
