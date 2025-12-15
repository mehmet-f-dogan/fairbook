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

type PriceLevel struct {
	price     Price
	orders    []*Order
	head      int
	exhausted bool
}

func newPriceLevel(price Price) *PriceLevel {
	return &PriceLevel{
		price:     price,
		orders:    make([]*Order, 0, 256),
		head:      0,
		exhausted: false,
	}
}

type OrderBook struct {
	Bids map[Price]*PriceLevel
	Asks map[Price]*PriceLevel

	BidPrices []Price
	AskPrices []Price
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		Bids:      make(map[Price]*PriceLevel, 32_768),
		Asks:      make(map[Price]*PriceLevel, 32_768),
		BidPrices: make([]Price, 0, 32_768),
		AskPrices: make([]Price, 0, 32_768),
	}
}
