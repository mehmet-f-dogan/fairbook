package engine

import (
	"bytes"
	"encoding/binary"
)

func (e *Engine) Snapshot() error {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, e.seq)
	return e.log.Append(Event{
		Type: EventSnapshot,
		Data: buf.Bytes(),
	})
}
