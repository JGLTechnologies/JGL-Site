package utils

import (
	"github.com/JGLTechnologies/GinRateLimit"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/imroc/req"
	"net/http"
	"os"
	"strconv"
	"time"
)

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
	client := http.Client{
		Timeout: time.Second * 5,
	}
	request := req.New()
	request.SetClient(&client)
	res, err := request.Get("https://api.pepy.tech/api/projects/" + project)
	if err != nil || res.Response().StatusCode != 200 {
		return "Not Found"
	}
	jsonErr := res.ToJSON(&data)
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
	client := http.Client{
		Timeout: time.Second * 5,
	}
	request := req.New()
	request.SetClient(&client)
	res, err := request.Get("https://api.npmjs.org/downloads/point/2020-1-1:" + date + "/" + project)
	if err != nil || res.Response().StatusCode != 200 {
		return "Not Found"
	}
	jsonErr := res.ToJSON(&data)
	if jsonErr != nil {
		return "Not Found"
	}
	return strconv.Itoa(int(data["downloads"].(float64)))
}

func GetGoLibDownloads(project string) string {
	request := req.New()
	var data map[string]interface{}
	client := http.Client{
		Timeout: time.Second * 5,
	}
	request.SetClient(&client)
	header := make(http.Header)
	header.Set("Authorization", "token "+os.Getenv("gh_token"))
	res, err := request.Get("https://api.github.com/repos/JGLTechnologies/"+project+"/traffic/clones?per=week", header)
	if err != nil || res.Response().StatusCode != 200 {
		return "Not Found"
	}
	jsonErr := res.ToJSON(&data)
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
