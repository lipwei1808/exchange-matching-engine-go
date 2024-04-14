package main

type PriceLevel []*Order

func (p PriceLevel) Len() int {
	return len(p)
}

func (p PriceLevel) Less(i, j int) bool {
	return p[i].price < p[j].price
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
