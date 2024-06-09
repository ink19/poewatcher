package poetrader

import (
	"context"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func (c *client) request(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.WithContext(ctx).Errorf("NewRequest fail, err: %v", err)
		return nil, err
	}

	if c.header != nil {
		req.Header = *c.header
	}

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.WithContext(ctx).Errorf("Request fail, err: %v", err)
		return nil, err
	}

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		log.WithContext(ctx).Errorf(" Read rsp, err: %v", err)
		return nil, err
	}

	return body, nil
}
