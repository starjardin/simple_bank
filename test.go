package main

import "fmt"

func main() {

	n := 10

	res := make(chan string)

	for i := range n {
		go func() {
			result := fmt.Sprintf("Hello world %d", i)

			res <- result
		}()
	}

	for range n {
		result := <-res
		fmt.Println(result)
	}

	fmt.Println("---------------------------------------------")

	for i := range n {
		result := fmt.Sprintf("Hello world %d", i)
		fmt.Println(result)
	}
}
