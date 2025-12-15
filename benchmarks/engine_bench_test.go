package benchmarks

import (
	"testing"

	"github.com/mehmet-f-dogan/fairbook/engine"
)

func BenchmarkMatch(b *testing.B) {
	e := engine.NewEngine(engine.NewOrderBook(), nil, engine.EmptyOnTradeFunc)

	var o engine.Order
	o.Side = engine.Buy
	o.Type = engine.Limit
	o.Price = 100
	o.Quantity = 10

	for i := 0; b.Loop(); i++ {
		o.ID = engine.OrderID(i)
		e.SubmitOrder(&o)
	}
}
