package engine

import (
	"bufio"
	"encoding/binary"
	"errors"
	"os"
)

type EventType uint8

const (
	EventAdd EventType = iota
	EventCancel
	EventTrade
	EventSnapshot
)

type Event struct {
	Type EventType
	Data []byte
}

type EventLog struct {
	file *os.File
	w    *bufio.Writer
}

func OpenEventLog(path string) (*EventLog, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return &EventLog{file: f, w: bufio.NewWriter(f)}, nil
}

func (l *EventLog) Append(e Event) error {
	binary.Write(l.w, binary.LittleEndian, e.Type)
	binary.Write(l.w, binary.LittleEndian, uint32(len(e.Data)))
	_, err := l.w.Write(e.Data)
	return err
}

type OnTradeHandler interface {
	ProcessTrade(BuyID OrderID,
		SellID OrderID,
		Qty Quantity,
		Price Price,
		Ts Ts)
}

type OnTradeFunc func(buy, sell OrderID, qty Quantity, price Price, ts Ts)

func EmptyOnTradeFunc(buy, sell OrderID, qty Quantity, price Price, ts Ts) {}

type Engine struct {
	book        *OrderBook
	log         *EventLog
	seq         uint64
	onTradeFunc OnTradeFunc
	orderIndex  map[OrderID]*Order
}

func NewEngine(book *OrderBook, log *EventLog, onTradeFunc OnTradeFunc) *Engine {
	return &Engine{
		book:        book,
		log:         log,
		onTradeFunc: onTradeFunc,
		orderIndex:  make(map[OrderID]*Order, 1<<16),
	}
}

func (e *Engine) SubmitOrder(o *Order) error {
	if o.Quantity <= 0 {
		return errors.New("invalid quantity")
	}

	e.seq++
	o.Ts = Ts(e.seq)

	if e.log != nil {
		e.log.Append(e.encodeAdd(o))
	}

	e.match(o)
	return nil
}

func (e *Engine) Cancel(id OrderID) error {
	if e.log != nil {
		var buf [8]byte
		binary.LittleEndian.PutUint64(buf[:], uint64(id))
		e.log.Append(Event{Type: EventCancel, Data: buf[:]})
	}
	return e.applyCancel(id)
}

func (e *Engine) applyCancel(id OrderID) error {
	o, ok := e.orderIndex[id]
	if !ok {
		return errors.New("order not found")
	}
	o.Canceled = true
	delete(e.orderIndex, id)
	return nil
}

func (e *Engine) match(o *Order) {
	if o.Side == Buy {
		e.matchBuy(o)
		if o.Quantity > 0 && o.Type == Limit {
			e.insertBid(o)
		}
		return
	}

	e.matchSell(o)
	if o.Quantity > 0 && o.Type == Limit {
		e.insertAsk(o)
	}
}

func (e *Engine) processTrade(o *Order, targetOrder *Order, price Price) {
	qty := min(o.Quantity, targetOrder.Quantity)
	ts := Ts(e.seq)

	e.onTradeFunc(o.ID, targetOrder.ID, qty, price, ts)

	if e.log != nil {
		e.log.Append(e.encodeTrade(o, targetOrder, qty, price, ts))
	}

	o.Quantity -= qty
	targetOrder.Quantity -= qty

	if o.Quantity == 0 {
		delete(e.orderIndex, o.ID)
	}

	if targetOrder.Quantity == 0 {
		delete(e.orderIndex, targetOrder.ID)
	}
}

func (e *Engine) matchBuy(o *Order) {
	for _, price := range e.book.askPrices {
		if price > o.Price || o.Quantity == 0 {
			break
		}

		lvl := e.book.asks[price]
		for lvl.head < len(lvl.orders) && o.Quantity > 0 {
			target := lvl.orders[lvl.head]
			if target.Canceled || target.Quantity == 0 {
				lvl.head++
				continue
			}
			e.processTrade(o, target, price)
			if target.Quantity == 0 {
				lvl.head++
			}
		}
		if lvl.head == len(lvl.orders) {
			lvl.exhausted = true
		}
	}
}

func (e *Engine) matchSell(o *Order) {
	for _, price := range e.book.bidPrices {
		if price < o.Price || o.Quantity == 0 {
			break
		}

		lvl := e.book.bids[price]
		for lvl.head < len(lvl.orders) && o.Quantity > 0 {
			target := lvl.orders[lvl.head]

			if target.Canceled || target.Quantity == 0 {
				lvl.head++
				continue
			}

			e.processTrade(o, target, price)

			if target.Quantity == 0 {
				lvl.head++
			}
		}
		if lvl.head == len(lvl.orders) {
			lvl.exhausted = true
		}
	}
}

func (e *Engine) CompactBook() {
	e.compactSide(e.book.bids, &e.book.bidPrices)
	e.compactSide(e.book.asks, &e.book.askPrices)
}

func (e *Engine) compactSide(
	levels map[Price]*priceLevel,
	prices *[]Price,
) {
	dst := (*prices)[:0]

	for _, p := range *prices {
		lvl := levels[p]
		if lvl.head == len(lvl.orders) {
			delete(levels, p)
			continue
		}
		dst = append(dst, p)
	}
	*prices = dst
}

func (e *Engine) encodeAdd(o *Order) Event {
	var buf [40]byte

	binary.LittleEndian.PutUint64(buf[0:], uint64(o.ID))
	binary.LittleEndian.PutUint64(buf[8:], uint64(o.Price))
	binary.LittleEndian.PutUint64(buf[16:], uint64(o.Quantity))
	binary.LittleEndian.PutUint64(buf[24:], uint64(o.Ts))
	buf[32] = byte(o.Side)
	buf[33] = byte(o.Type)

	return Event{
		Type: EventAdd,
		Data: buf[:],
	}
}

func (e *Engine) encodeTrade(o *Order, targetOrder *Order, qty Quantity, price Price, ts Ts) Event {
	var buf [40]byte

	if o.Side == Buy {
		binary.LittleEndian.PutUint64(buf[0:], uint64(o.ID))
		binary.LittleEndian.PutUint64(buf[8:], uint64(targetOrder.ID))
	} else {
		binary.LittleEndian.PutUint64(buf[0:], uint64(targetOrder.ID))
		binary.LittleEndian.PutUint64(buf[8:], uint64(o.ID))
	}

	binary.LittleEndian.PutUint64(buf[16:], uint64(qty))
	binary.LittleEndian.PutUint64(buf[24:], uint64(price))
	binary.LittleEndian.PutUint64(buf[32:], uint64(ts))
	return Event{
		Type: EventTrade,
		Data: buf[:],
	}
}

func min(a, b Quantity) Quantity {
	if a < b {
		return a
	}
	return b
}
