package poetrader

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	poeInfoRequest = "https://poe.game.qq.com/api/trade/fetch/%s?query=%s"
)

type GetInfoRes struct {
	Result []*PoeGood `json:"result"`
}

func (c *client) GetInfo(ctx context.Context, searchID string, goodID string) (*PoeGood, error) {
	err := rateLimit.Wait(ctx)
	if err != nil {
		log.WithContext(ctx).Errorf("Rate limit fail, err: %v", err)
		return nil, err
	}
	reqURL := fmt.Sprintf(poeInfoRequest, goodID, searchID)
	rspBody, err := c.request(ctx, reqURL)
	if err != nil {
		log.WithContext(ctx).Errorf("Request fail, err: %v", err)
		return nil, err
	}
	log.WithContext(ctx).Debugf("Request %s: %s", goodID, string(rspBody))
	res := &GetInfoRes{}
	err = json.Unmarshal(rspBody, res)
	if err != nil {
		log.WithContext(ctx).Errorf("Unmarshal fail, err: %v", err)
		return nil, err
	}
	if len(res.Result) == 0 {
		log.WithContext(ctx).Errorf("Unmarshal fail, err: %v", err)
		return &PoeGood{
			ID: goodID,
		}, nil
	}
	return res.Result[0], err
}

func (c *client) BatchGetInfo(ctx context.Context, searchID string, goodIDs []string) ([]*PoeGood, error) {
	err := rateLimit.Wait(ctx)
	if err != nil {
		log.WithContext(ctx).Errorf("Rate limit fail, err: %v", err)
		return nil, err
	}
	reqURL := fmt.Sprintf(poeInfoRequest, strings.Join(goodIDs, ","), searchID)
	rspBody, err := c.request(ctx, reqURL)
	if err != nil {
		log.WithContext(ctx).Errorf("Request fail, err: %v", err)
		return nil, err
	}
	res := &GetInfoRes{}
	err = json.Unmarshal(rspBody, res)
	if err != nil {
		log.WithContext(ctx).Errorf("Unmarshal fail, err: %v", err)
		return nil, err
	}
	if len(res.Result) == 0 {
		log.WithContext(ctx).Errorf("Empty result")
		return nil, fmt.Errorf("empty result")
	}
	return res.Result, err
}
