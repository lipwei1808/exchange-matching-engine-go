package main

import (
	"container/heap"
	"context"
)

type Prices struct {
	prices    map[uint32]*PriceLevel
	inputChan chan *Order
	oppChan   chan *Order
}

func NewPrices(ctx context.Context, oppChan chan *Order) *Prices {
	p := &Prices{
		prices:    make(map[uint32]*PriceLevel),
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

func (p *Prices) Execute(ctx context.Context, oppOrder *Order) {
	// check for valid matches
	matches := p.Match(oppOrder)

	if matches == oppOrder.count {
		return
	}

	// check oppChan if order is to be added
	select {
	case <-ctx.Done():
		break
	case o := <-p.oppChan:
		// add order to heap
		p.Add(o)

		// re-execute current oppOrder
		p.Execute(ctx, oppOrder)

	// if no order in oppChan, send order to add.
	case p.oppChan <- oppOrder:
		return
	}
}

// Attempts to match opposing order with resting orders on the heap
// @param oppOrder order of the opposing type
// @returns number of successful quantity matches
func (p *Prices) Match(oppOrder *Order) uint32 {
	return 0
}

// Adds an order of the same type to the heap
func (p *Prices) Add(o *Order) {
	_, exists := p.prices[o.price]
	if !exists {
		os := make(PriceLevel, 0)
		p.prices[o.price] = &os
	}
	pl, _ := p.prices[o.price]
	heap.Push(pl, o)
}
