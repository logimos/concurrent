package main

import (
	"context"
	"fmt"

	"github.com/logimos/concurrent"
)

func main() {
	ctx := context.Background()

	input := make(chan int)
	pipeline := concurrent.NewPipeline[int](ctx)

	// Multiply by 2
	pipeline.AddStage(concurrent.Map(func(n int) int {
		return n * 2
	}))

	// Filter even numbers
	pipeline.AddStage(concurrent.Filter(func(n int) bool {
		return n%2 == 0
	}))

	output := pipeline.Run(input)

	// Send data
	go func() {
		for i := 1; i <= 10; i++ {
			input <- i
		}
		close(input)
	}()

	// Process results
	for result := range output {
		fmt.Println(result)
	}

	pipeline.Close()
}

