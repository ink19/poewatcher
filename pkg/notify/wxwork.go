package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type wxWorkClient struct {
	reqURI string
}

type wxMsgType string

const (
	wxMsgTypeText wxMsgType = "text"
)

type wxWorkTextMsg struct {
	Content string `json:"content"`
}

type wxWorkMsg struct {
	MsgType wxMsgType      `json:"msgtype"`
	Text    *wxWorkTextMsg `json:"text,omitempty"`
}

func NewWxWork(reqURI string) Client {
	return &wxWorkClient{
		reqURI: reqURI,
	}
}

func (c *wxWorkClient) SendTextMsg(ctx context.Context, msg string) error {
	body, err := c.packTextMsg(ctx, msg)
	if err != nil {
		log.WithContext(ctx).Errorf("Pack Text msg fail, err: %v", err)
		return err
	}
	err = c.invoke(ctx, body)
	if err != nil {
		log.WithContext(ctx).Errorf("Invoke fail, err: %v", err)
		return err
	}
	return nil
}

func (c *wxWorkClient) packTextMsg(ctx context.Context, msg string) ([]byte, error) {
	sMsg := &wxWorkMsg{
		MsgType: wxMsgTypeText,
		Text: &wxWorkTextMsg{
			Content: msg,
		},
	}

	r, err := json.Marshal(sMsg)
	if err != nil {
		log.WithContext(ctx).Errorf("Marshal fail, err: %v", err)
		return nil, err
	}
	return r, nil
}

func (c *wxWorkClient) invoke(ctx context.Context, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.reqURI, bytes.NewReader([]byte(body)))
	if err != nil {
		log.WithContext(ctx).Debugf("Request WxWork fail, err: %v", err)
		return err
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.WithContext(ctx).Debugf("Do WxWork fail, err: %v", err)
		return err
	}
	rspBody, err := io.ReadAll(rsp.Body)
	if err != nil {
		log.WithContext(ctx).Debugf("Read WxWork rsp fail, err: %v", err)
		return err
	}
	log.WithContext(ctx).Debugf("rspBody: %s", string(rspBody))
	return nil
}
