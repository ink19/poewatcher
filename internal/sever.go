package internal

import (
	"context"
	"encoding/base64"

	"github.com/ink19/poewatcher/pkg/poetrader"
	"github.com/sirupsen/logrus"
)

const seasonID = "S25赛季"
const searchID = "A3zgF5"
const cookieStr = ""

func RunServer() {
	ctx := context.Background()
	poeClient := poetrader.New(seasonID, cookieStr)
	ch, err := poeClient.Watch(ctx, searchID)
	if err != nil {
		logrus.WithContext(ctx).Errorf("Watch search fail, err: %v", err)
		return
	}
	for good := range ch {
		logrus.WithContext(ctx).Debugf("goodID: %s", good.ID)
		good, err := poeClient.GetInfo(ctx, searchID, good.ID)
		if err != nil {
			logrus.WithContext(ctx).Errorf("GetInfo fail, err: %v", err)
			continue
		}
		logrus.WithContext(ctx).Debugf("GetInfo succ, good: %v", good)
		desc, err := base64.StdEncoding.DecodeString(good.Item.Extended.DescText)
		if err != nil {
			logrus.WithContext(ctx).Errorf("GetDesc fail, err: %v", err)
			continue
		}
		logrus.WithContext(ctx).Debugf("%s", desc)
	}
}
