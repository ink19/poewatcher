package poetrader

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
)

const watchURLFormat = "wss://poe.game.qq.com/api/trade/live/%s/%s"

type wsMessage struct {
	messageType int
	message string
}

type wsRecvMsg struct {
	Auth *bool `json:"auth,omitempty"`
	New []string `json:"new,omitempty"`
}

func (c *client) Watch(ctx context.Context, searchID string) (<-chan *PoeGood, error) {
	log.WithContext(ctx).Debugf("Enter Watch")
	ch := make(chan *PoeGood)
	watchURL := fmt.Sprintf(watchURLFormat, c.seasonID, searchID)
	header := &http.Header{}
	if c.header != nil {
		header = c.header
	}
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, watchURL, *header)
	if err != nil {
		log.WithContext(ctx).Errorf("WS connect fail, err: %v", err)
		return nil, err
	}
	log.WithContext(ctx).Debugf("BeginWatch")
	msgChan, err := c.readWSConn(ctx, conn)
	if err != nil {
		log.WithContext(ctx).Errorf("ReadWsConn fail, err: %v", err)
		return nil, err
	}
	go func ()  {
		for {
			hasDone := false
			select {
			case <- ctx.Done():
				hasDone = true
			case msg := <- msgChan:
				if msg == nil {
					break
				}
				log.WithContext(ctx).Debugf("Recv msg: %s", msg.message)
				recvMsg := &wsRecvMsg{}
				err := json.Unmarshal([]byte(msg.message), recvMsg)
				if err != nil {
					log.WithContext(ctx).Errorf("Unmarshal fail, err: %v", err)
					break
				}
				if recvMsg.Auth != nil {
					if !*recvMsg.Auth {
						log.WithContext(ctx).Errorf("Auth fail")
						hasDone = true
					} else {
						log.WithContext(ctx).Debugf("Auth succ")
					}
					break
				}
				for _, goodID := range recvMsg.New {
					ch <- &PoeGood{
						ID: goodID,
					}
				}
			}

			if hasDone {
				log.WithContext(ctx).Debugf("Ctx done")
				conn.Close()
				break
			}
		}

		for remainMsg := range msgChan {
			_ = remainMsg
			log.WithContext(ctx).Debugf("Recv msg: %s", remainMsg.message)
			ch <- &PoeGood{}
		}
	}()
	return ch, nil
}

func (c *client) readWSConn(ctx context.Context, conn *websocket.Conn) (chan *wsMessage, error) {
	msgChan := make(chan *wsMessage, 10)
	go func ()  {
		defer close(msgChan)
		for {
			mt, ms, err := conn.ReadMessage()
			if err != nil {
				log.WithContext(ctx).Errorf("ReadMessage fail, err: %v", err)
				return
			}
			msgChan <- &wsMessage{
				messageType: mt,
				message: string(ms),
			}
		}
	}()
	return msgChan, nil
}
