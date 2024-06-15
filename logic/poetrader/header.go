package poetrader

import "net/http"

const (
	simUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"
)

func GetSimHeader(cookieStr string) *http.Header {
	header := http.Header{}
	header.Add("Cookie", cookieStr)
	header.Add("Host", "poe.game.qq.com")
	header.Add("Pragma", "no-cache")
	header.Add("Cache-Control", "no-cache")
	header.Add("User-Agent", simUA)
	header.Add("Origin", "https://poe.game.qq.com")
	header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	return &header
}
