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
	defer func() {
		close(p.inputChan)
	}()

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
	log.Printf("[prices.Cancel (%c)] BEFORE DELETE %d: %s", p.pricesType, o.orderId, p.prices.ToString())
	f := p.prices.Delete(o.orderId)
	log.Printf("[prices.Cancel (%c)] AFTER DELETE: %s", p.pricesType, p.prices.ToString())

	outputOrderDeleted(o.ToInput(), f, GetCurrentTimestamp())
}

func (p *Prices) Execute(ctx context.Context, req PricesRequest) bool {
	oppOrder := req.order
	// check if valid order type
	if oppOrder.orderType == p.pricesType {
		return false
	}

	// check for valid matches
	done := p.Match(oppOrder)

	if done {
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
func (p *Prices) Match(incoming *Order) bool {
	for incoming.count > 0 && len(p.prices) > 0 && p.IsMatchable(incoming) {
		log.Printf("[prices.Match (%c)] orderid: %d, %s", p.pricesType, incoming.orderId, p.prices.ToString())

		resting := p.prices[0]
		resting.executionId++
		MatchOrders(resting, incoming)

		if resting.count == 0 {
			heap.Pop(&p.prices)
		}
	}

	return incoming.count == 0
}

func MatchOrders(resting, incoming *Order) uint32 {
	qty := uint32(math.Min(float64(resting.count), float64(incoming.count)))
	log.Printf("[prices.MatchOrders] qty: %d, resting cnt: %d, incomnig cnt: %d\n", qty, resting.count, incoming.count)
	resting.Fill(qty)
	incoming.Fill(qty)

	outputOrderExecuted(
		resting.orderId,
		incoming.orderId,
		resting.executionId,
		resting.price,
		qty,
		GetCurrentTimestamp(),
	)
	return qty
}

// Adds an order of the same type to the heap
func (p *Prices) Add(o *Order) {
	log.Printf("[prices.Add (%c)] order: %d\n", p.pricesType, o.orderId)
	o.timestamp = GetCurrentTimestamp()
	if o.orderType != p.pricesType {
		return
	}
	log.Printf("[prices.Add (%c)] BEFORE ADD %d: %s", p.pricesType, o.orderId, p.prices.ToString())
	heap.Push(&p.prices, o)
	log.Printf("[prices.Add (%c)] AFTER ADD %d: %s", p.pricesType, o.orderId, p.prices.ToString())

	outputOrderAdded(o.ToInput(), o.timestamp)
}
