# fairbook

A deterministic, allocation-free order matching engine designed for low-latency trading systems.

## Overview

`fairbook` is a single-process, in-memory order matching engine with strict price–time priority. It is not a full exchange. There is no networking layer, persistence service, or risk engine. The goal of this project is to focus entirely on the core matching problem and do it well.

The engine is designed to be embedded inside a larger trading or exchange system. Each instance is intended to serve a single symbol.

Key design goals:

- Deterministic behavior
- Allocation-free matching path
- Predictable latency under load
- Replayable event log
- Simple, inspectable data structures

---

## Features

- Price–time priority (FIFO at each price level)
- Zero allocations in the matching hot path
- In-memory order book
- Event logging for deterministic replay
- Cancel support in O(1) via order indexing
- No goroutines, no locks, no background threads
- Suitable for simulation, backtesting, and live systems

---

## Non-Goals

This project intentionally does **not** include:

- Networking or APIs
- Risk checks or margining
- Persistence beyond the event log
- Multi-symbol routing
- Concurrency inside the engine

Those concerns belong outside the matching engine and are left to the integrator.

---

## Architecture

### One engine per symbol

Each `Engine` instance owns:

- One `OrderBook`
- One sequence counter
- One optional event log

This mirrors how real exchanges and trading venues scale. Symbols are routed to engines upstream. Engines remain single-threaded and deterministic.

---

### Order book structure

The order book uses:

- Maps from `Price → priceLevel`
- Compact price slices (`bidPrices`, `askPrices`)
- FIFO order queues per price level
- Head indices instead of linked lists

This design avoids pointer chasing, minimizes cache misses, and keeps cancellation and matching predictable.

---

## Performance

Benchmark run on:

- CPU: Intel i7-1255U
- OS: Linux
- Go version: 1.25

```bash
BenchmarkMatch-12 7,551,435 179.4 ns/op 0 allocs/op
```

Notes:

- Zero allocations in the matching path
- Single-threaded execution
- No unsafe code
- No GC pressure during steady-state operation

These numbers are intended as a sanity check, not a marketing claim. Real systems should benchmark under realistic workloads.

---

## Determinism and Replay

All state transitions can be captured as events:

- Order add
- Order cancel
- Trade execution

Given the same event stream, the engine will deterministically rebuild the same book state. This is useful for:

- Crash recovery
- Backtesting
- Simulation
- Auditing

Replay logic intentionally mirrors the live execution path.

---

## Usage

```go

import "github.com/mehmet-f-dogan/fairbook/engine"

func main(){
    book := engine.NewOrderBook()
    eng := engine.NewEngine(book, nil, engine.EmptyOnTradeFunc)

    err := eng.SubmitOrder(&engine.Order{
        ID:       1,
        Side:     engine.Buy,
        Type:     engine.Limit,
        Price:    100,
        Quantity: 10,
    })
}

```

Each engine instance should be driven by a single goroutine.

### Who this is for

- Low-latency trading systems

- Exchange or ATS prototypes

- Backtesting engines

- Systems engineers interested in market microstructure

- Engineers who care about determinism and performance
