package main

import (
	"container/heap"
	"context"
	"errors"
	"fmt"
	"math"
)

type Prices struct {
	prices     PriceLevel
	pricesType inputType
	inputChan  chan *Order
	oppChan    chan *Order
}

func NewPrices(ctx context.Context, oppChan chan *Order, pricesType inputType) (*Prices, error) {
	if pricesType == inputCancel {
		return nil, errors.New("only allow buy or sell inputType")
	}

	p := &Prices{
		prices:    make(PriceLevel, 0),
		inputChan: make(chan *Order),
		oppChan:   oppChan,
	}

	go p.pricesWorker(ctx)

	return p, nil
}

func (p *Prices) pricesWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case o := <-p.inputChan:
			p.Execute(ctx, o)
		}
	}
}

func (p *Prices) HandleOrder(o *Order) {
	p.inputChan <- o
}

func (p *Prices) Execute(ctx context.Context, oppOrder *Order) {
	// check if valid order type
	if oppOrder.orderType == p.pricesType {
		return
	}

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
		return
	// if no order in oppChan, send order to add.
	case p.oppChan <- oppOrder:
		return
	}
}

func (p *Prices) IsMatchable(oppOrder *Order) bool {
	if oppOrder.orderType == inputBuy && oppOrder.price < p.prices[0].price {
		return false
	}

	if oppOrder.orderType == inputSell && oppOrder.price > p.prices[0].price {
		return false
	}

	return true
}

// Attempts to match opposing order with resting orders on the heap
// @param oppOrder order of the opposing type
// @returns number of successful quantity matches
func (p *Prices) Match(incoming *Order) uint32 {
	matched := uint32(0)
	for incoming.count > 0 {
		// check if any valid orders to match
		if len(p.prices) == 0 {
			break
		}

		if !p.IsMatchable(incoming) {
			break
		}

		resting := p.prices[0]
		resting.executionId++
		MatchOrders(resting, incoming)

		if resting.count == 0 {
			heap.Pop(&p.prices)
		}
	}

	return matched
}

func MatchOrders(resting, incoming *Order) {
	qty := uint32(math.Min(float64(resting.count), float64(incoming.count)))
	if resting.count == incoming.count {
		resting.Fill(qty)
		incoming.Fill(qty)
	} else if resting.count > incoming.count {
		resting.Fill(incoming.count)
		incoming.Fill(incoming.count)
	} else {
		resting.Fill(resting.count)
		incoming.Fill(resting.count)
	}

	outputOrderExecuted(
		resting.orderId,
		incoming.orderId,
		resting.executionId,
		resting.price,
		qty,
		GetCurrentTimestamp(),
	)
}

// Adds an order of the same type to the heap
func (p *Prices) Add(o *Order) {
	fmt.Printf("[Add] %c order: %d\n", p.pricesType, o.orderId)
	if o.orderType != p.pricesType {
		return
	}

	heap.Push(&p.prices, o)
	outputOrderAdded(o.ToInput(), GetCurrentTimestamp())
}
