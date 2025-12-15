package engine

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const (
	maxEventSize = 256 // hard cap
)

func Replay(logPath string) (*Engine, error) {
	f, err := os.Open(logPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	book := NewOrderBook()

	engine := NewEngine(book, nil, EmptyOnTradeFunc)

	reader := bufio.NewReaderSize(f, 1<<20)

	var (
		typ  EventType
		size uint32
	)

	for {
		if err := binary.Read(reader, binary.LittleEndian, &typ); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("replay: read event type: %w", err)
		}

		if err := binary.Read(reader, binary.LittleEndian, &size); err != nil {
			return nil, fmt.Errorf("replay: read event size: %w", err)
		}

		if size > maxEventSize {
			return nil, fmt.Errorf("replay: event size %d exceeds limit", size)
		}

		var data [maxEventSize]byte
		buf := data[:size]

		if _, err := io.ReadFull(reader, buf); err != nil {
			return nil, fmt.Errorf("replay: read payload: %w", err)
		}

		switch typ {

		case EventAdd:
			o := decodeAdd(buf)

			if uint64(o.Ts) > engine.seq {
				engine.seq = uint64(o.Ts)
			}

			engine.match(o)

		case EventCancel:
			id := OrderID(binary.LittleEndian.Uint64(buf))
			_ = engine.applyCancel(id)

		case EventSnapshot:
			if len(buf) >= 8 {
				seq := binary.LittleEndian.Uint64(buf)
				engine.seq = seq
			}

		case EventTrade:
			// Derived event: ignored intentionally

		default:
			return nil, fmt.Errorf("replay: unknown event type %d", typ)
		}
	}

	return engine, nil
}

func decodeAdd(buf []byte) *Order {
	return &Order{
		ID:       OrderID(binary.LittleEndian.Uint64(buf[0:])),
		Price:    Price(binary.LittleEndian.Uint64(buf[8:])),
		Quantity: Quantity(binary.LittleEndian.Uint64(buf[16:])),
		Ts:       Ts(binary.LittleEndian.Uint64(buf[24:])),
		Side:     Side(buf[32]),
		Type:     OrderType(buf[33]),
	}
}
