package utils

import (
	"github.com/JGLTechnologies/GinRateLimit"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/imroc/req/v3"
	"os"
	"strconv"
	"time"
)

var client = req.C().SetTimeout(time.Second * 5)

func StartsWith(s string, sw string) bool {
	swLen := len(sw)
	sLen := len(s)
	if swLen > sLen {
		return false
	} else if s[:swLen] == sw {
		return true
	} else {
		return false
	}
}

func GetPythonLibDownloads(project string) string {
	var data map[string]interface{}
	res, err := client.R().Get("https://api.pepy.tech/api/projects/" + project)
	if err != nil || res.IsError() {
		return "Not Found"
	}
	jsonErr := res.UnmarshalJson(&data)
	if jsonErr != nil {
		return "Not Found"
	}
	return strconv.Itoa(int(data["total_downloads"].(float64)))
}

func GetNPMLibDownloads(project string) string {
	var date string
	date += strconv.Itoa(time.Now().Year())
	date += "-" + strconv.Itoa(int(time.Now().Month()))
	date += "-" + strconv.Itoa(time.Now().Day())
	var data map[string]interface{}
	res, err := client.R().Get("https://api.npmjs.org/downloads/point/2020-1-1:" + date + "/" + project)
	if err != nil || res.IsError() {
		return "Not Found"
	}
	jsonErr := res.UnmarshalJson(&data)
	if jsonErr != nil {
		return "Not Found"
	}
	return strconv.Itoa(int(data["downloads"].(float64)))
}

func GetGoLibDownloads(project string) string {
	var data map[string]interface{}
	res, err := client.R().SetHeader("Authorization", "token "+os.Getenv("gh_token")).Get("https://api.github.com/repos/JGLTechnologies/" + project + "/traffic/clones?per=week")
	if err != nil || res.IsError() {
		return "Not Found"
	}
	jsonErr := res.UnmarshalJson(&data)
	if jsonErr != nil {
		return "Not Found"
	}
	return strconv.Itoa(int(data["count"].(float64)))
}

func GetMW(rate time.Duration, limit int) func(c *gin.Context) {
	return GinRateLimit.RateLimiter(func(c *gin.Context) string {
		return c.ClientIP() + c.FullPath()
	}, func(c *gin.Context, remaining time.Duration) {
		c.String(429, "Too many requests. Try again in "+remaining.String())
	}, GinRateLimit.InMemoryStore(rate, limit))
}

func GetWS(c *gin.Context, upGrader websocket.Upgrader) *websocket.Conn {
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(err)
	} else {
		c.Set("ws", ws)
		c.Next()
	}
	return ws
}
