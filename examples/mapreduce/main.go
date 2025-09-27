package main

import (
	"context"
	"fmt"

	"github.com/logimos/concurrent"
)

func main() {
	ctx := context.Background()

	data := []int{1, 2, 3, 4, 5, 6}
	out, err := concurrent.MapConcurrent(ctx, data, 3, func(ctx context.Context, v int) (int, error) {
		return v * v, nil
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(out) // [1 4 9 16 25 36]
}
