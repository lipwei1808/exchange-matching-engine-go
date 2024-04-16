package main

import "context"

type OrderBookInput struct {
	order     *Order
	orderType inputType
}

type OrderBook struct {
	bids      *Prices
	asks      *Prices
	inputChan chan *OrderBookInput
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
		inputChan: make(chan *OrderBookInput),
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
				ob.asks.HandleOrder(o.order)
				break
			case inputSell:
				ob.bids.HandleOrder(o.order)
				break
			default:
				if o.orderType == inputBuy {
					ob.bids.HandleOrder(o.order)
				} else {
					ob.asks.HandleOrder(o.order)

				}
				break
			}
			return
		}
	}
}

func (ob *OrderBook) HandleOrder(order *Order) {
	ob.inputChan <- order
}
