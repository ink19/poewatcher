package dao

import "context"

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
	UpdateRecord(ctx context.Context, record *Record) error
	GetRecord(ctx context.Context, id int64) (*Record, error)
	ListRecords(ctx context.Context) ([]*Record, error)
	DeleteRecord(ctx context.Context, id int64) error
}

type client struct {}

func NewClient() Client {
	return &client{}
}

func (c *client) AddRecord(ctx context.Context, record *Record) error {
	return nil
}

func (c *client) UpdateRecord(ctx context.Context, record *Record) error {
	return nil
}

func (c *client) GetRecord(ctx context.Context, id int64) (*Record, error) {
	return nil, nil
}

func (c *client) ListRecords(ctx context.Context) ([]*Record, error) {
	return nil, nil
}

func (c *client) DeleteRecord(ctx context.Context, id int64) error {
	return nil
}
