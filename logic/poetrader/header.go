package poetrader

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

const (
	simUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
)

func GetSimHeader(cookieStr string) *http.Header {
	header := http.Header{}
	logrus.Debugf("cookie: %s", cookieStr)
	header.Add("Cookie", cookieStr)
	header.Add("Host", "poe.game.qq.com")
	header.Add("Pragma", "no-cache")
	header.Add("Cache-Control", "no-cache")
	header.Add("User-Agent", simUA)
	header.Add("Origin", "https://poe.game.qq.com")
	header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	return &header
}
