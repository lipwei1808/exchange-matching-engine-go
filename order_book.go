package main

import "context"

type OrderBook struct {
	bids      *Prices
	asks      *Prices
	inputChan chan *Order
}

func NewOrderBook(ctx context.Context) *OrderBook {
	oppChan := make(chan *Order)
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
				ob.asks.HandleOrder(o)
				break
			case inputSell:
				ob.bids.HandleOrder(o)
				break
			default:
				ob.asks.HandleOrder(o)
				ob.bids.HandleOrder(o)
				break
			}
			return
		}
	}
}

func (ob *OrderBook) HandleOrder(order *Order) {
	ob.inputChan <- order
}
