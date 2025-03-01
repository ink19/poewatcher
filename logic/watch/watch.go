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
	c poetrader.Client
	done context.CancelFunc

	lock sync.Locker
	wg sync.WaitGroup
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
	w.done = done

	logrus.WithContext(ctx).Debugf("Begin Run, record: %v", w.record)

	if err := w.initRecord(ctx); err != nil {
		logrus.WithContext(ctx).Errorf("initRecord fail, err: %v", err)
		return err
	}

	w.wg.Add(1)
	go func ()  {
		defer w.wg.Done()
		_ = w.WatchRecord(ctx)
	}()
	return nil
}

func (w *watcher) Record() *dao.Record {
	return w.record
}

func (w *watcher) Stop() {
	w.lock.Lock()
	defer w.lock.Unlock()
	
	_ = w.c.Stop(w.ctx)
	w.record.Status = dao.RecordStatusPending
	err := dao.NewClient().UpdateRecordStatus(w.ctx, w.record.ID, dao.RecordStatusPending)
	if err != nil {
		logrus.WithContext(w.ctx).Errorf("update record status fail, err: %v", err)
	}
}

func (w *watcher) Delete() {
	w.lock.Lock()
	defer w.lock.Unlock()

	_ = w.c.Stop(w.ctx)
	if err := dao.NewClient().DeleteRecord(w.ctx, w.record.ID); err != nil {
		logrus.WithContext(w.ctx).Errorf("delete record fail, err: %v", err)
	}
}

func (w *watcher) WatchRecord(ctx context.Context) error {
	poeClient := poetrader.New(w.record.SeasonID, w.record.Cookie)
	w.c = poeClient
	
	ch, err := poeClient.Watch(ctx, w.record.SearchID)
	if err != nil {
		logrus.WithContext(ctx).Errorf("Watch search fail, err: %v", err)
		return nil
	}

	// 使用新的ctx，不影响原来的ctx
	ctx = context.Background()
	notifyClient := notify.NewWxWork(config.Get().Notify.URL)
	for good := range ch {
		logrus.WithContext(ctx).Debugf("goodID: %s", good.ID)
		good, err := poeClient.GetInfo(ctx, w.record.SearchID, good.ID)
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

func (w *watcher) initRecord(ctx context.Context) error {
	if w.record.ID == 0 {
		logrus.WithContext(ctx).Debugf("w.record.ID is 0, add to sql")
		w.record.Status = dao.RecordStatusRunning
		err := dao.NewClient().AddRecord(ctx, w.record)
		if err != nil {
			logrus.WithContext(ctx).Errorf("AddRecord fail, err: %v", err)
			return err
		}
	} else {
		if w.record.Status == dao.RecordStatusRunning {
			logrus.WithContext(ctx).Debugf("record %d Status is %d, skip", w.record.ID, w.record.Status)
			return nil
		}
		logrus.WithContext(ctx).Debugf("w.record.ID is %d, status is %d", w.record.ID, w.record.Status)
		w.record.Status = dao.RecordStatusRunning
		err := dao.NewClient().UpdateRecordStatus(ctx, w.record.ID, w.record.Status)
		if err != nil {
			logrus.WithContext(ctx).Errorf("UpdateRecordStatus fail, err: %v", err)
			return err
		}
	}
	return nil
}
