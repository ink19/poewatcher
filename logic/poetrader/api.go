package poetrader

import (
	"context"
	"net/http"
	"sync"

	"github.com/ink19/poewatcher/config"
	"golang.org/x/time/rate"
)

type PoeItemExtended struct {
	DescText string `json:"text"`
}

type PoeItem struct {
	Extended PoeItemExtended `json:"extended"`
}

type PoeListing struct {
	Price PoePrice `json:"price"`
}

type PoePrice struct {
	Type     string `json:"type"`
	Amount   int    `json:"amount"`
	Currency string `json:"divine"`
}

type PoeGood struct {
	ID      string     `json:"id"`
	Listing PoeListing `json:"listing"`
	Item    PoeItem    `json:"Item"`
}

type Client interface {
	GetInfo(ctx context.Context, searchID string, goodID string) (*PoeGood, error)
	Watch(ctx context.Context, searchID string) (<-chan *PoeGood, error)
}

var (
	dbOnce = &sync.Once{}
	rateLimit rate.Limiter
)

func New(seasonID string, cookies string) Client {
	dbOnce.Do(func() {
		rateLimit = *rate.NewLimiter(rate.Limit(config.Get().Poe.RateLimit), config.Get().Poe.RateLimit)
	})

	return &client{
		cookies:  cookies,
		header:   GetSimHeader(cookies),
		seasonID: seasonID,
	}
}

type client struct {
	cookies  string
	seasonID string

	header *http.Header
}
