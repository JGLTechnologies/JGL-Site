package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"time"
)

var BotStatus = timeout.New(timeout.WithTimeout(time.Second), timeout.WithResponse(botStatusTimeout), timeout.WithHandler(botStatusResponse))
var BotInfo = timeout.New(timeout.WithTimeout(time.Second), timeout.WithResponse(botInfoTimeout), timeout.WithHandler(botInfoResponse))

func botStatusResponse(c *gin.Context) {
	client := http.Client{
		Timeout: time.Second,
	}
	_, err := client.Get("https://jglbotapi.us/status")
	if err != nil {
		c.Abort()
	} else {
		c.JSON(200, gin.H{"online": true})
	}
}

func botStatusTimeout(c *gin.Context) {
	c.JSON(200, gin.H{"online": false})
}

func botInfoResponse(c *gin.Context) {
	client := http.Client{
		Timeout: time.Second,
	}
	var data map[string]interface{}
	res, err := client.Get("https://jglbotapi.us/info")
	if err != nil {
		c.Abort()
	} else {
		defer res.Body.Close()
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("%s", err)})
		}
		json.Unmarshal(bodyBytes, &data)
		c.JSON(200, data)
	}
}

func botInfoTimeout(c *gin.Context) {
	c.JSON(200, gin.H{"guilds": "Not Found", "cogs": "Not Found", "shards": "Not Found", "size": gin.H{"gb": "Not Found", "mb": "Not Found", "kb": "Not Found"}, "ping": "Not Found"})
}
