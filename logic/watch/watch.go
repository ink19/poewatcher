package watch

import (
	"context"
	"encoding/base64"
	"sync"

	"github.com/ink19/poewatcher/config"
	"github.com/ink19/poewatcher/logic/dao"
	"github.com/ink19/poewatcher/logic/poetrader"
	"github.com/ink19/poewatcher/pkg/notify"
	"github.com/sirupsen/logrus"
)

type Watcher interface {
	Run() error
	Stop()
	Delete()
	Record() *dao.Record
}

type watcher struct {
	record *dao.Record
	ctx context.Context
	cancel context.CancelFunc

	lock sync.Locker
}

func New(r *dao.Record) Watcher {
	return &watcher{
		record: r,
		lock: &sync.Mutex{},
	}
}

func (w *watcher) Run() error {
	w.lock.Lock()
	defer w.lock.Unlock()

	ctx, done := context.WithCancel(context.Background())
	w.ctx = ctx
	w.cancel = done

	if err := initRecord(ctx, w.record); err != nil {
		logrus.WithContext(ctx).Errorf("initRecord fail, err: %v", err)
		return err
	}

	go func ()  {
		_ = WatchRecord(ctx, w.record)
	}()
	return nil
}

func (w *watcher) Record() *dao.Record {
	return w.record
}

func (w *watcher) Stop() {
	w.lock.Lock()
	defer w.lock.Unlock()

	err := dao.NewClient().UpdateRecordStatus(w.ctx, w.record.ID, dao.RecordStatusPending)
	if err != nil {
		logrus.WithContext(w.ctx).Errorf("update record status fail, err: %v", err)
	}
	w.cancel()
}

func (w *watcher) Delete() {
	w.lock.Lock()
	defer w.lock.Unlock()
	
	if err := dao.NewClient().DeleteRecord(w.ctx, w.record.ID); err != nil {
		logrus.WithContext(w.ctx).Errorf("delete record fail, err: %v", err)
	}
	w.cancel()
}

func WatchRecord(ctx context.Context, record *dao.Record) error {
	poeClient := poetrader.New(record.SeasonID, record.Cookie)
	ch, err := poeClient.Watch(ctx, record.SearchID)
	if err != nil {
		logrus.WithContext(ctx).Errorf("Watch search fail, err: %v", err)
		return nil
	}

	// 使用新的ctx，不影响原来的ctx
	ctx = context.Background()
	notifyClient := notify.NewWxWork(config.Get().Notify.URL)
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
		err = notifyClient.SendTextMsg(ctx, string(desc))
		if err != nil {
			logrus.WithContext(ctx).Errorf("SendTextMsg fail, err: %v", err)
		}
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
