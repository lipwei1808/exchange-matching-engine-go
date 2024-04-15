package main

type Order struct {
	orderType   inputType
	orderId     uint32
	price       uint32
	count       uint32
	instrument  string
	executionId uint32
	timestamp   int64
}

func (o Order) ToInput() input {
	return input{
		orderType:  o.orderType,
		orderId:    o.orderId,
		price:      o.price,
		count:      o.count,
		instrument: o.instrument,
	}
}

func (o *Order) Fill(qty uint32) {
	if qty > o.count {
		qty = o.count
	}

	o.count -= qty
}
