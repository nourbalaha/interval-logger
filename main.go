package main

import (
	"context"
	"fmt"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// 1 streamer
// 2 logger

type Data struct {
	mut       sync.Mutex
	count     int64
	timestamp time.Time
}

func getData(ctx context.Context, data *Data) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			data.mut.Lock()
			data.count += 1
			data.timestamp = time.Now()
			data.mut.Unlock()
			<-time.After(100 * time.Millisecond)
		}
	}
}

func logger(ctx context.Context, data *Data, complete chan bool) {
	defer func() {
		fmt.Println("completed logger")
	}()

	for {
		select {
		case <-ctx.Done():
			complete <- true
			return
		default:
			data.mut.Lock()
			count := data.count
			timestamp := data.timestamp
			data.mut.Unlock()
			fmt.Printf("[%v] %v at %v\n", time.Now().Format(time.StampMilli), count, timestamp.Format(time.StampMilli))
			<-time.After(time.Second)
		}
	}
}

func main() {

	signalCtx, cancelSignal := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer cancelSignal()

	ctx, cancel := context.WithTimeout(signalCtx, time.Second*10)
	defer cancel()
	someData := &Data{}
	completeCh := make(chan bool, 1)

	go getData(ctx, someData)
	go logger(ctx, someData, completeCh)

	fmt.Println("waiting")
	<-signalCtx.Done()
	fmt.Println("Main done")
	valueGot := <-completeCh
	fmt.Printf("got this complete value: %v\n", valueGot)
}
