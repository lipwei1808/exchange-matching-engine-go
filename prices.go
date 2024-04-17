package main

import (
	"container/heap"
	"context"
	"errors"
	"log"
	"math"
)

type PricesRequest struct {
	order  *Order
	output chan interface{}
}

func (p PricesRequest) Done() {
	log.Printf("[prices.Done] orderid: %d", p.order.orderId)
	p.output <- struct{}{}
}

type Prices struct {
	prices     PriceLevel
	pricesType inputType
	inputChan  chan PricesRequest
	oppChan    chan PricesRequest
}

func NewPrices(ctx context.Context, oppChan chan PricesRequest, pricesType inputType) (*Prices, error) {
	if pricesType == inputCancel {
		return nil, errors.New("only allow buy or sell inputType")
	}

	p := &Prices{
		pricesType: pricesType,
		prices:     make(PriceLevel, 0),
		inputChan:  make(chan PricesRequest),
		oppChan:    oppChan,
	}

	go p.pricesWorker(ctx)

	return p, nil
}

func (p *Prices) pricesWorker(ctx context.Context) {
	for {
		log.Printf("[prices.pricesWorker (%c)] NEW LOOP heaplen: %d \n", p.pricesType, len(p.prices))
		select {
		case <-ctx.Done():
			return
		case req := <-p.oppChan:
			log.Printf("[prices.pricesWorker.oppChan (%c)] orderid: %d\n", p.pricesType, req.order.orderId)
			p.Add(req.order)

			req.Done()
			break

		case req := <-p.inputChan:
			log.Printf("[prices.pricesWorker.inputChan (%c)] orderid: %d, ordertype: %c\n", p.pricesType, req.order.orderId, req.order.orderType)

			switch req.order.orderType {
			case inputCancel:
				p.Cancel(req.order)
				req.Done()
				break
			default:
				done := p.Execute(ctx, req)
				if done {
					req.Done()
				}
			}
		}
		log.Printf("[prices.pricesWorker (%c)] END LOOP\n", p.pricesType)
	}
}

func (p *Prices) HandleOrder(o PricesRequest) {
	log.Printf("[prices.HandleOrder (%c)] orderid: %d, heaplen (%d)\n", p.pricesType, o.order.orderId, len(p.prices))
	p.inputChan <- o
}

func (p *Prices) Cancel(o *Order) {
	log.Printf("[prices.Cancel (%c)] orderid: %d, orderType: %c, heaplen: %d\n", p.pricesType, o.orderId, o.orderType, len(p.prices))
	f := false
	for i, d := range p.prices {
		log.Printf("[prices.Cancel-loop %d (%c)], orderid: %d\n", i, p.pricesType, d.orderId)
		if d.orderId == o.orderId {
			f = true
			p.prices[i], p.prices[len(p.prices)-1] = p.prices[len(p.prices)-1], p.prices[i]
			p.prices = p.prices[:len(p.prices)-1]
			break
		}
	}

	heap.Init(&p.prices)
	outputOrderDeleted(o.ToInput(), f, GetCurrentTimestamp())
}

func (p *Prices) Execute(ctx context.Context, req PricesRequest) bool {
	oppOrder := req.order
	// check if valid order type
	if oppOrder.orderType == p.pricesType {
		return false
	}

	// check for valid matches
	matches := p.Match(oppOrder)

	if matches == oppOrder.count {
		return true
	}

	// check oppChan if order is to be added
	select {
	case <-ctx.Done():
		break
	case r := <-p.oppChan:
		// perform operation on order
		p.Add(r.order)
		r.Done()

		// re-execute current oppOrder
		return p.Execute(ctx, req)
	// if no order in oppChan, send order to add.
	case p.oppChan <- req:
		return false
	}

	return false
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
	resting.Fill(qty)
	incoming.Fill(qty)

	log.Printf("[prices.MatchOrders] qty: %d, resting cnt: %d, incomnig cnt: %d\n", qty, resting.count, incoming.count)
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
	log.Printf("[prices.Add (%c)] order: %d\n", p.pricesType, o.orderId)
	if o.orderType != p.pricesType {
		return
	}

	heap.Push(&p.prices, o)
	log.Printf("[prices.Add (%c)] new heaplen: %d\n", p.pricesType, len(p.prices))
	outputOrderAdded(o.ToInput(), GetCurrentTimestamp())
}
