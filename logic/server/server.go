package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/ink19/poewatcher/logic/dao"
	"github.com/ink19/poewatcher/logic/watch"
	"github.com/sirupsen/logrus"
)

type recordStorage struct {
	data map[int64]watch.Watcher
	lock sync.RWMutex
}

func (s *recordStorage) add(r watch.Watcher) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data[r.Record().ID] = r
}

func (s *recordStorage) get(id int64) (watch.Watcher, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	r, ok := s.data[id]
	return r, ok
}

func (s *recordStorage) delete(id int64) (watch.Watcher, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	r, ok := s.data[id]
	if ok {
		delete(s.data, id)
	}

	return r, ok
}

type Server interface {
	Run() error
	Stop() error
}

type server struct{
	recordStorage

	service *http.Server
}

func New() Server {
	return &server{
		recordStorage: recordStorage{
			data: make(map[int64]watch.Watcher),
			lock: sync.RWMutex{},
		},
	}
}

func (s *server) Run() error {
	router := gin.Default()
	router.POST("/add", s.add)
	router.GET("/delete", s.delete)
	router.GET("/get", s.get)
	router.GET("/list", s.list)
	router.GET("/pause", s.pause)
	router.GET("/start", s.start)

	records, err := dao.NewClient().ListRecords(context.Background())
	if err != nil {
		logrus.WithError(err).Error("failed to get records from dao")
		return err
	}
	for _, r := range records {
		w := watch.New(r)
		if err = w.Run(); err != nil {
			logrus.WithError(err).Error("failed to run watcher")
			continue
		}
		s.recordStorage.add(w)
	}

	s.service = &http.Server{Addr: ":8080", Handler: router}

	return s.service.ListenAndServe()
}

func (s *server) add(ctx *gin.Context) {
	req, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		logrus.WithError(err).Error("failed to read request body")
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	record := &dao.Record{}
	if err = json.Unmarshal(req, record); err != nil {
		logrus.WithError(err).Error("failed to unmarshal request body")
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}
	
	w := watch.New(record)
	if err = w.Run(); err != nil {
		logrus.WithError(err).Error("failed to run watcher")
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	s.recordStorage.add(w)
	ctx.JSON(200, gin.H{"id": w.Record().ID})
}

func (s *server) start(ctx *gin.Context) {
	idStr, ok := ctx.GetQuery("id")
	if !ok {
		logrus.Error("failed to get id")
		ctx.JSON(400, gin.H{"error": "No id"})
		return
	}
	
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logrus.WithError(err).Error("failed to parse id")
		ctx.JSON(400, gin.H{"error": "Invalid id"})
		return
	}

	w, ok := s.recordStorage.get(id)
	if !ok {
		record, err := dao.NewClient().GetRecord(ctx, id)
		if err != nil {
			logrus.WithError(err).Error("failed to get record from dao")
			ctx.JSON(400, gin.H{"error": err.Error()})
			return
		}
		w = watch.New(record)
	}

	if err = w.Run(); err != nil {
		logrus.WithError(err).Error("failed to run watcher")
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{"id": w.Record().ID})
}

func (s *server) delete(ctx *gin.Context) {
	idStr, ok := ctx.GetQuery("id")
	if !ok {
		logrus.Error("failed to get id")
		ctx.JSON(400, gin.H{"error": "No id"})
		return
	}
	
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logrus.WithError(err).Error("failed to parse id")
		ctx.JSON(400, gin.H{"error": "Invalid id"})
		return
	}

	w, ok := s.recordStorage.delete(id)
	if !ok {
		ctx.JSON(404, gin.H{"error": "not found"})
		return
	}

	w.Delete()
	ctx.JSON(200, gin.H{"id": w.Record().ID})
}

func (s *server) get(ctx *gin.Context) {
	idStr, ok := ctx.GetQuery("id")
	if !ok {
		logrus.Error("failed to get id")
		ctx.JSON(400, gin.H{"error": "No id"})
		return
	}
	
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logrus.WithError(err).Error("failed to parse id")
		ctx.JSON(400, gin.H{"error": "Invalid id"})
		return
	}

	w, ok := s.recordStorage.get(id)
	if !ok {
		ctx.JSON(404, gin.H{"error": "not found"})
		return
	}

	ctx.JSON(200, w.Record())
}

func (s *server) list(ctx *gin.Context) {
	records, err := dao.NewClient().ListRecords(ctx)
	if err != nil {
		logrus.WithError(err).Error("failed to list records")
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, records)
}

func (s *server) pause(ctx *gin.Context) {
	idStr, ok := ctx.GetQuery("id")
	if !ok {
		logrus.Error("failed to get id")
		ctx.JSON(400, gin.H{"error": "No id"})
		return
	}
	
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		logrus.WithError(err).Error("failed to parse id")
		ctx.JSON(400, gin.H{"error": "Invalid id"})
		return
	}

	w, ok := s.recordStorage.get(id)
	if !ok {
		ctx.JSON(404, gin.H{"error": "not found"})
		return
	}

	w.Stop()
	ctx.JSON(200, gin.H{"id": w.Record().ID})
}

func (s *server) Stop() error {
	for _, w := range s.recordStorage.data {
		w.Stop()
	}

	return s.service.Shutdown(context.Background())
}
