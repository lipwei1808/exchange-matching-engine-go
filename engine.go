package main

import "C"
import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type MainChanRequest struct {
	input  *input
	output chan *OrderBook
}

type Engine struct {
	instruments map[string]*OrderBook
	mainChan    chan *MainChanRequest
}

func NewEngine(ctx context.Context) *Engine {
	mainChan := make(chan *MainChanRequest)
	e := &Engine{
		mainChan:    mainChan,
		instruments: make(map[string]*OrderBook),
	}

	go e.engineWorker(ctx)

	return e
}

func (e *Engine) engineWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case p := <-e.mainChan:
			ob := e.GetOrderBook(ctx, p.input.instrument)
			p.output <- ob
		}
	}
}

func (e *Engine) accept(ctx context.Context, conn net.Conn) {
	go func() {
		<-ctx.Done()
		conn.Close()
	}()
	go handleConn(conn)
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	for {
		in, err := readInput(conn)
		if err != nil {
			if err != io.EOF {
				_, _ = fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			}
			return
		}
		switch in.orderType {
		case inputCancel:
			fmt.Fprintf(os.Stderr, "Got cancel ID: %v\n", in.orderId)
			outputOrderDeleted(in, true, GetCurrentTimestamp())
		default:
			fmt.Fprintf(os.Stderr, "Got order: %c %v x %v @ %v ID: %v\n",
				in.orderType, in.instrument, in.count, in.price, in.orderId)
			outputOrderAdded(in, GetCurrentTimestamp())
		}
		outputOrderExecuted(123, 124, 1, 2000, 10, GetCurrentTimestamp())
	}
}

func (e *Engine) GetOrderBook(ctx context.Context, instrument string) *OrderBook {
	_, exists := e.instruments[instrument]
	if !exists {
		e.instruments[instrument] = NewOrderBook(ctx)
	}

	return e.instruments[instrument]
}

func GetCurrentTimestamp() int64 {
	return time.Now().UnixNano()
}
