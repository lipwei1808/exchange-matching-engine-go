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
