package blockchain

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// maticPriceCacheTTL bounds how often the MATIC/PLN rate is refetched — a gas
// cost preview doesn't need a rate fresher than a few minutes, and this keeps
// well under CoinGecko's free-tier rate limit.
const maticPriceCacheTTL = 5 * time.Minute

const maticPriceURL = "https://api.coingecko.com/api/v3/simple/price?ids=matic-network&vs_currencies=pln"

var maticPriceCache struct {
	mu        sync.Mutex
	pln       float64
	fetchedAt time.Time
}

// MaticPLNPrice returns the current MATIC->PLN exchange rate from CoinGecko,
// caching it for maticPriceCacheTTL.
func MaticPLNPrice(ctx context.Context) (float64, error) {
	maticPriceCache.mu.Lock()
	if time.Since(maticPriceCache.fetchedAt) < maticPriceCacheTTL {
		price := maticPriceCache.pln
		maticPriceCache.mu.Unlock()
		return price, nil
	}
	maticPriceCache.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, maticPriceURL, nil)
	if err != nil {
		return 0, fmt.Errorf("build price request: %w", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("fetch matic price: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("fetch matic price: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		MaticNetwork struct {
			PLN float64 `json:"pln"`
		} `json:"matic-network"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, fmt.Errorf("decode matic price: %w", err)
	}

	maticPriceCache.mu.Lock()
	maticPriceCache.pln = body.MaticNetwork.PLN
	maticPriceCache.fetchedAt = time.Now()
	maticPriceCache.mu.Unlock()

	return body.MaticNetwork.PLN, nil
}
