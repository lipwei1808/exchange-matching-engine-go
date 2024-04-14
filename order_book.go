package main

import "context"

type OrderBook struct {
	bids      *Prices
	asks      *Prices
	inputChan chan *Order
}

func NewOrderBook(ctx context.Context) *OrderBook {
	oppChan := make(chan *Order)
	ob := OrderBook{
		bids:      NewPrices(ctx, oppChan),
		asks:      NewPrices(ctx, oppChan),
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
				ob.asks.Execute(ctx, o)
				break
			case inputSell:
				ob.bids.Execute(ctx, o)
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
