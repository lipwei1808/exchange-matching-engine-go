package main

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
