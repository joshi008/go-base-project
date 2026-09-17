package main

import (
	"fmt"
	"strconv"
)

func odd(n int, ch chan string, oddFinal chan bool, evensyncy chan bool, oddsyncy chan bool) {
	for i := 1; i <= n; i++ {
		if i%2 != 0 {
			s := "Odd : " + strconv.Itoa(i)
			<-oddsyncy
			fmt.Println(s)
			evensyncy <- false
		}
	}
	oddFinal <- true
}

func even(n int, ch chan string, evenFinal chan bool, evensyncy chan bool, oddsyncy chan bool) {
	for i := 1; i <= n; i++ {
		if i%2 == 0 {
			s := "Even : " + strconv.Itoa(i)
			<-evensyncy
			fmt.Println(s)
			oddsyncy <- true
		}
	}
	evenFinal <- true
}

func main() {
	n := 10
	ch := make(chan string, n)
	evenSync := make(chan bool, n)
	oddSync := make(chan bool, n)

	evenFinal := make(chan bool)
	oddFinal := make(chan bool)
	fmt.Println("Starting of the program")

	go even(n, ch, evenFinal, evenSync, oddSync)
	go odd(n, ch, oddFinal, evenSync, oddSync)

	oddSync <- true

	p1 := <-evenFinal
	p2 := <-oddFinal
	fmt.Println(p1, p2)
}
