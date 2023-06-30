package main

import (
	"context"
	"example/golang-graceful-shutdown/pool"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	run(ctx)
}

func run(ctx context.Context) {
	pool := pool.StartNewPool(ctx, 25)

	// add a few jobs initially
	for i := 0; i < 100; i++ {
		pool.AddNewJob(func() {
			fmt.Println("I am doing hard work")
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(2000)+1000))
		})
	}
	// add new jobs every 100ms
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Millisecond * 100):
				pool.AddNewJob(func() {
					fmt.Println("I am doing hard work later")
					time.Sleep(time.Millisecond * time.Duration(rand.Intn(2000)+1000))
				})
			}
		}
	}()

	// wait for the pool to finish
	pool.Wait()
}
