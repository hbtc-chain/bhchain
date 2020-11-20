package orderbook

import (
	"fmt"
	"sort"

	sdk "github.com/hbtc-chain/bhchain/types"
	"github.com/hbtc-chain/bhchain/x/openswap/types"
)

type Manager struct {
	markets        map[string]*Market
	marketKeys     []string
	openswapKeeper OpenswapKeeper
}

func NewManager(openswapKeeper OpenswapKeeper) *Manager {
	return &Manager{
		markets:        make(map[string]*Market),
		openswapKeeper: openswapKeeper,
	}
}

func (m *Manager) Init(ctx sdk.Context) {
	m.openswapKeeper.IteratorAllUnfinishedOrder(ctx, m.AddOrder)
}

func (m *Manager) AddOrder(order *types.Order) {
	key := m.formatKey(order.DexID, order.BaseSymbol, order.QuoteSymbol)
	market, exist := m.markets[key]
	if !exist {
		market = m.addMarket(order.DexID, order.BaseSymbol, order.QuoteSymbol)
	}
	market.AddOrder(order)
}

func (m *Manager) DelOrder(order *types.Order) {
	key := m.formatKey(order.DexID, order.BaseSymbol, order.QuoteSymbol)
	market, exist := m.markets[key]
	if !exist {
		market = m.addMarket(order.DexID, order.BaseSymbol, order.QuoteSymbol)
	}
	market.DelOrder(order)
}

func (m *Manager) addMarket(dexID uint32, baseSymbol, quoteSymbol sdk.Symbol) *Market {
	market := NewMarket(dexID, baseSymbol, quoteSymbol)
	key := m.formatKey(dexID, baseSymbol, quoteSymbol)
	m.markets[key] = market
	m.marketKeys = append(m.marketKeys, key)
	sort.Strings(m.marketKeys)
	return market
}

func (m *Manager) GetAllOrders(dexID uint32, baseSymbol, quoteSymbol sdk.Symbol) ([]*types.Order, []*types.Order) {
	market, exist := m.markets[m.formatKey(dexID, baseSymbol, quoteSymbol)]
	if !exist {
		return nil, nil
	}
	return market.GetAllOrders()
}

func (m *Manager) GetExpiredOrders(ctx sdk.Context) []*types.Order {
	var expired []*types.Order

	for _, marketKey := range m.marketKeys {
		market := m.markets[marketKey]
		orders := market.GetExpiredOrders(ctx)
		expired = append(expired, orders...)
	}

	return expired
}

func (m *Manager) GetMarkets() []*Market {
	markets := make([]*Market, len(m.marketKeys))
	for i, key := range m.marketKeys {
		markets[i] = m.markets[key]
	}
	return markets
}

func (m *Manager) formatKey(dexID uint32, baseSymbol, quoteSymbol sdk.Symbol) string {
	return fmt.Sprintf("%d-%s-%s", dexID, baseSymbol, quoteSymbol)
}
