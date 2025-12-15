package engine

type Price uint64
type Quantity uint64
type OrderID uint64
type Ts uint64

type Side uint8

const (
	Buy Side = iota
	Sell
)

type OrderType uint8

const (
	Limit OrderType = iota
	Market
)

type Order struct {
	ID       OrderID
	Side     Side
	Type     OrderType
	Price    Price
	Quantity Quantity
	Ts       Ts
	Canceled bool
}

type priceLevel struct {
	price     Price
	orders    []*Order
	head      int
	exhausted bool
}

func newPriceLevel(price Price) *priceLevel {
	return &priceLevel{
		price:     price,
		orders:    make([]*Order, 0, 256),
		head:      0,
		exhausted: false,
	}
}

type OrderBook struct {
	bids map[Price]*priceLevel
	asks map[Price]*priceLevel

	bidPrices []Price
	askPrices []Price
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		bids:      make(map[Price]*priceLevel, 32_768),
		asks:      make(map[Price]*priceLevel, 32_768),
		bidPrices: make([]Price, 0, 32_768),
		askPrices: make([]Price, 0, 32_768),
	}
}
