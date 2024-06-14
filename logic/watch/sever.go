package watch

import (
	"context"
	"encoding/base64"

	"github.com/ink19/poewatcher/logic/dao"
	"github.com/ink19/poewatcher/pkg/poetrader"
	"github.com/sirupsen/logrus"
)

func WatchRecord(ctx context.Context, record *dao.Record) error {
	if err := initRecord(ctx, record); err != nil {
		logrus.WithContext(ctx).Errorf("initRecord fail, err: %v", err)
		return err
	}
	poeClient := poetrader.New(record.SeasonID, record.Cookie)
	ch, err := poeClient.Watch(ctx, record.SearchID)
	if err != nil {
		logrus.WithContext(ctx).Errorf("Watch search fail, err: %v", err)
		return nil
	}
	for good := range ch {
		logrus.WithContext(ctx).Debugf("goodID: %s", good.ID)
		good, err := poeClient.GetInfo(ctx, record.SearchID, good.ID)
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
	return nil
}

func initRecord(ctx context.Context, record *dao.Record) error {
	if record.ID == 0 {
		logrus.WithContext(ctx).Errorf("record.ID is 0, add to sql")
		record.Status = dao.RecordStatusRunning
		err := dao.NewClient().AddRecord(ctx, record)
		if err != nil {
			logrus.WithContext(ctx).Errorf("AddRecord fail, err: %v", err)
			return err
		}
	} else {
		if record.Status == dao.RecordStatusRunning {
			logrus.WithContext(ctx).Debugf("record %d Status is %d, skip", record.ID, record.Status)
			return nil
		}
		logrus.WithContext(ctx).Debugf("record.ID is %d, status is %d", record.ID, record.Status)
		record.Status = dao.RecordStatusRunning
		err := dao.NewClient().UpdateRecordStatus(ctx, record.ID, record.Status)
		if err != nil {
			logrus.WithContext(ctx).Errorf("UpdateRecordStatus fail, err: %v", err)
			return err
		}
	}
	return nil
}
