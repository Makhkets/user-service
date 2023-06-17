package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	doSomething(ctx)
}

func doSomething(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Истекло время")
			return
		default:
			go func() {
				time.Sleep(2 * time.Second)
				fmt.Println("Работаю")
				return
			}()
		}
	}
}
