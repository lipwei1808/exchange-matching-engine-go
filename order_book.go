package main

import "context"

type OrderBook struct {
	bids      *Prices
	asks      *Prices
	inputChan chan *Order
}

func NewOrderBook(ctx context.Context) *OrderBook {
	ob := OrderBook{
		bids:      NewPrices(),
		asks:      NewPrices(),
		inputChan: make(chan *Order),
	}

	go ob.orderBookWorker(ctx)

	return &ob
}

func (ob *OrderBook) orderBookWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case o := <-ob.inputChan:
			switch o.orderType {
			case inputBuy:
				break
			case inputSell:
				break
			default:
				break
			}
			return
		}
	}
}

func (ob *OrderBook) handleOrder(order *Order) {
	ob.inputChan <- order
}
