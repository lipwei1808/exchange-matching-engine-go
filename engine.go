package main

import "C"
import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

type MainChanRequest struct {
	instrument string
	output     chan *OrderBook
}

type Engine struct {
	instruments map[string]*OrderBook
	mainChan    chan MainChanRequest
}

func NewEngine(ctx context.Context) *Engine {
	e := &Engine{
		mainChan:    make(chan MainChanRequest),
		instruments: make(map[string]*OrderBook),
	}

	go e.engineWorker(ctx)

	return e
}

func (e *Engine) engineWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			close(e.mainChan)
			return
		case p := <-e.mainChan:
			ob := e.GetOrderBook(ctx, p.instrument)
			p.output <- ob
		}
	}
}

func (e *Engine) accept(ctx context.Context, conn net.Conn) {
	go func() {
		<-ctx.Done()
		conn.Close()
	}()
	go e.handleConn(conn)
}

func (e *Engine) handleConn(conn net.Conn) {
	defer conn.Close()
	orders := make(map[uint32]*Order)

	for {
		output := make(chan interface{})
		in, err := readInput(conn)
		if err != nil {
			if err != io.EOF {
				_, _ = fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			}
			return
		}
		o := Order{
			orderType:   in.orderType,
			orderId:     in.orderId,
			price:       in.price,
			count:       in.count,
			instrument:  in.instrument,
			executionId: 0,
		}
		switch in.orderType {
		case inputCancel:
			fmt.Fprintf(os.Stderr, "Got cancel ID: %v\n", in.orderId)
			oo, exists := orders[in.orderId]
			if !exists {
				outputOrderDeleted(in, false, GetCurrentTimestamp())
				continue
			}

			ob := e.RequestOrderBook(oo.instrument)
			ob.HandleOrder(OrderBookRequest{order: &o, orderType: oo.orderType, output: output})
			break
		default:
			fmt.Fprintf(os.Stderr, "Got order: %c %v x %v @ %v ID: %v\n",
				in.orderType, in.instrument, in.count, in.price, in.orderId)
			orders[o.orderId] = &o

			ob := e.RequestOrderBook(o.instrument)
			ob.HandleOrder(OrderBookRequest{order: &o, orderType: in.orderType, output: output})
		}

		<-output
		close(output)
	}
}

func (e *Engine) RequestOrderBook(i string) *OrderBook {
	output := make(chan *OrderBook)
	req := MainChanRequest{
		instrument: i,
		output:     output,
	}
	e.mainChan <- req

	ob := <-output
	close(output)
	return ob
}

func (e *Engine) GetOrderBook(ctx context.Context, instrument string) *OrderBook {
	_, exists := e.instruments[instrument]
	if !exists {
		log.Printf("[engine.GetOrderBook] instrument: %s not found\n", instrument)
		e.instruments[instrument] = NewOrderBook(ctx)
	}

	return e.instruments[instrument]
}

func GetCurrentTimestamp() int64 {
	return time.Now().UnixNano()
}
