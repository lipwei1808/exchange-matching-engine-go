package main

import (
	"container/heap"
	"fmt"
	"log"
)

type PriceLevel []*Order

func (p PriceLevel) Len() int {
	return len(p)
}

// Ordering based on price, timestamp then id
func (p PriceLevel) Less(i, j int) bool {
	if p[i].price == p[j].price {
		if p[i].timestamp == p[j].timestamp {
			return p[i].orderId < p[j].orderId
		}

		return p[i].timestamp < p[j].timestamp
	}

	if p[i].orderType == inputSell {
		return p[i].price < p[j].price
	}

	return p[j].price < p[i].price
}

func (p PriceLevel) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p *PriceLevel) Push(x interface{}) {
	*p = append(*p, x.(*Order))
}

func (p *PriceLevel) Pop() any {
	old := *p
	n := len(old)
	x := old[n-1]
	*p = old[0 : n-1]
	return x
}

func (p *PriceLevel) Delete(id uint32) bool {
	f := false
	old := *p
	for i, d := range old {
		if d.orderId == id {
			f = true
			old[i], old[len(old)-1] = old[len(old)-1], old[i]
			*p = old[:len(old)-1]
			break
		}
	}

	heap.Init(p)
	return f
}

func (p PriceLevel) Print() {
	s := ""
	for i, d := range p {
		de := "->"
		if i == len(p)-1 {
			de = ""
		}
		s += fmt.Sprintf("(%d, p:%d, c:%d)%s", d.orderId, d.price, d.count, de)
	}
	log.Printf("%s\n", s)
}
