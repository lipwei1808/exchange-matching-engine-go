package main

import "context"

type Prices struct {
}

func NewPrices() *Prices {
	return &Prices{}
}

func (p *Prices) pricesWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		}
	}
}
