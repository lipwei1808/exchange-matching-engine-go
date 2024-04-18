package main

import (
	"context"
	"log"
)

type OrderBookRequest struct {
	order     *Order
	orderType inputType
	output    chan interface{}
}

type OrderBook struct {
	bids      *Prices
	asks      *Prices
	inputChan chan OrderBookRequest
}

// Creates a new order book and initialises the two buy and
// sell heaps. Responsible for initialising the orderbook
// goroutine and the communication channel (oppChan)
// between the 2 heaps.
func NewOrderBook(ctx context.Context) *OrderBook {
	oppChan := make(chan PricesRequest)
	bids, err := NewPrices(ctx, oppChan, inputBuy)
	if err != nil {
		panic(err)
	}

	asks, err := NewPrices(ctx, oppChan, inputSell)
	if err != nil {
		panic(err)
	}

	ob := OrderBook{
		bids:      bids,
		asks:      asks,
		inputChan: make(chan OrderBookRequest),
	}

	go ob.orderBookWorker(ctx, oppChan)

	return &ob
}

func (ob *OrderBook) orderBookWorker(ctx context.Context, oppChan chan PricesRequest) {
	defer func() {
		close(oppChan)
		close(ob.inputChan)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case req := <-ob.inputChan:
			log.Printf("[order_book.orderBookWorker] order: %d, type: %c\n", req.order.orderId, req.order.orderType)

			r := PricesRequest{order: req.order, output: req.output}
			switch req.order.orderType {
			case inputBuy:
				ob.asks.HandleOrder(r)
				break
			case inputSell:
				ob.bids.HandleOrder(r)
				break
			default:
				if req.orderType == inputBuy {
					ob.bids.HandleOrder(r)
				} else {
					ob.asks.HandleOrder(r)
				}
				break
			}
		}
	}
}

func (ob *OrderBook) HandleOrder(req OrderBookRequest) {
	log.Printf("[order_book.HandleOrder] order: %d\n", req.order.orderId)
	ob.inputChan <- req
}
