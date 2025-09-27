package main

import (
	"context"
	"fmt"
	"time"

	"github.com/logimos/concurrent"
)

func main() {
	ctx := context.Background()

	jobs := make(chan int)
	pool := concurrent.NewPool(3, func(ctx context.Context, v int) (string, error) {
		time.Sleep(15 * time.Millisecond)
		return fmt.Sprintf("processed-%d", v), nil
	})
	results := pool.Run(ctx, jobs)

	go func() {
		for i := 0; i < 8; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	for r := range results {
		fmt.Println(r)
	}
}
