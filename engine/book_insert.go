package engine

func (e *Engine) insertBid(o *Order) {
	lvl, ok := e.book.Bids[o.Price]
	if ok {
		lvl.orders = append(lvl.orders, o)
		e.orderIndex[o.ID] = o
		return
	}

	lvl = newPriceLevel(o.Price)
	lvl.orders = append(lvl.orders, o)

	e.book.Bids[o.Price] = lvl
	e.book.BidPrices = insertPriceDesc(e.book.BidPrices, o.Price)
	e.orderIndex[o.ID] = o
}

func (e *Engine) insertAsk(o *Order) {
	lvl, ok := e.book.Asks[o.Price]
	if ok {
		lvl.orders = append(lvl.orders, o)
		e.orderIndex[o.ID] = o
		return
	}

	lvl = newPriceLevel(o.Price)
	lvl.orders = append(lvl.orders, o)

	e.book.Asks[o.Price] = lvl
	e.book.AskPrices = insertPriceAsc(e.book.AskPrices, o.Price)
	e.orderIndex[o.ID] = o
}

func insertPriceDesc(prices []Price, p Price) []Price {
	i := 0
	for i < len(prices) && prices[i] > p {
		i++
	}

	if i < len(prices) && prices[i] == p {
		return prices
	}

	prices = append(prices, 0)
	copy(prices[i+1:], prices[i:])
	prices[i] = p
	return prices
}

func insertPriceAsc(prices []Price, p Price) []Price {
	i := 0
	for i < len(prices) && prices[i] < p {
		i++
	}

	if i < len(prices) && prices[i] == p {
		return prices
	}

	prices = append(prices, 0)
	copy(prices[i+1:], prices[i:])
	prices[i] = p
	return prices
}
