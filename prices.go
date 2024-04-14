package main

import "context"

type Prices struct {
	prices    map[uint32]*[]Order
	inputChan chan *Order
	oppChan   chan *Order
}

func NewPrices(ctx context.Context, oppChan chan *Order) *Prices {
	p := &Prices{
		prices:    make(map[uint32]*[]Order),
		inputChan: make(chan *Order),
		oppChan:   oppChan,
	}

	go p.pricesWorker(ctx)

	return p
}

func (p *Prices) pricesWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		}
	}
}

func (p *Prices) execute(ctx context.Context, oppOrder *Order) {
	// check for valid matches
	matches := match(oppOrder)

	if matches == oppOrder.count {
		return
	}

	// check oppChan if order is to be added
	select {
	case <-ctx.Done():
		break
	case o := <-p.oppChan:
		// add order to heap
		add(o)

		// re-execute current oppOrder
		p.execute(ctx, oppOrder)

	// if no order in oppChan, send order to add.
	case p.oppChan <- oppOrder:
		return
	}
}

// Attempts to match opposing order with resting orders on the heap
// @param oppOrder order of the opposing type
// @returns number of successful quantity matches
func match(oppOrder *Order) uint32 {
	return 0
}

// Adds an order of the same type to the heap
func add(order *Order) {

}
