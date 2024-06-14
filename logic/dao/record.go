package dao

import (
	"context"
	"database/sql"
	"sync"

	"github.com/ink19/poewatcher/config"
	"github.com/sirupsen/logrus"
)

type RecordStatusEnum int

const (
	RecordStatusNone RecordStatusEnum = iota
	RecordStatusRunning
	RecordStatusPending
	RecordStatusError
)

type Record struct {
	ID       int64            `json:"id"`
	SeasonID string           `json:"season_id"`
	SearchID string           `json:"search_id"`
	Cookie   string           `json:"cookie"`
	Status   RecordStatusEnum `json:"status"`
}

type Client interface {
	AddRecord(ctx context.Context, record *Record) error
	UpdateRecordStatus(ctx context.Context, id int64, status RecordStatusEnum) error
	GetRecord(ctx context.Context, id int64) (*Record, error)
	ListRecords(ctx context.Context) ([]*Record, error)
	DeleteRecord(ctx context.Context, id int64) error
}

type client struct {}

var (
	dbOnce = &sync.Once{}
	dbHandler *sql.DB
)

func NewClient() Client {
	dbOnce.Do(func() {
		var err error
		dbHandler, err = sql.Open("sqlite3", config.Get().DB.Path + ":locked.sqlite?cache=shared")
		if err != nil {
			logrus.Errorf("open db error: %s", err)
			panic(err)
		}
		dbHandler.SetMaxOpenConns(1)
	})
	return &client{}
}

func (c *client) AddRecord(ctx context.Context, record *Record) error {
	_, err := dbHandler.Exec("INSERT INTO record (season_id, search_id, cookie, status) VALUES (?, ?, ?, ?)", record.SeasonID, record.SearchID, record.Cookie, record.Status)
	if err != nil {
		return err
	}
	return nil
}

func (c *client) UpdateRecordStatus(ctx context.Context, id int64, status RecordStatusEnum) error {
	_, err := dbHandler.Exec("UPDATE record SET status = ? WHERE id = ?", status, id)
	return err
}

func (c *client) GetRecord(ctx context.Context, id int64) (*Record, error) {
	rows, err := dbHandler.Query("SELECT id, season_id, search_id, cookie, status FROM record WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		record := &Record{}
		err := rows.Scan(&record.ID, &record.SeasonID, &record.SearchID, &record.Cookie, &record.Status)
		if err != nil {
			return nil, err
		}
		return record, nil
	}
	return nil, nil
}

func (c *client) ListRecords(ctx context.Context) ([]*Record, error) {
	rows, err := dbHandler.Query("SELECT id, season_id, search_id, cookie, status FROM record")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []*Record
	for rows.Next() {
		record := &Record{}
		err := rows.Scan(&record.ID, &record.SeasonID, &record.SearchID, &record.Cookie, &record.Status)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (c *client) DeleteRecord(ctx context.Context, id int64) error {
	_, err := dbHandler.Exec("DELETE FROM record WHERE id = ?", id)
	return err
}
