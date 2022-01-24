package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"time"
)

func BotStatus(c *gin.Context) {
	client := http.Client{
		Timeout: time.Second,
	}
	_, err := client.Get("https://jglbotapi.us/status")
	if err != nil {
		c.AbortWithStatusJSON(200, gin.H{"online": false})
	} else {
		c.JSON(200, gin.H{"online": true})
	}
}

func BotInfo(c *gin.Context) {
	client := http.Client{
		Timeout: time.Second,
	}
	var data map[string]interface{}
	res, err := client.Get("https://jglbotapi.us/info")
	if err != nil {
		c.AbortWithStatusJSON(200, gin.H{"guilds": "Not Found", "cogs": "Not Found", "shards": "Not Found", "size": gin.H{"gb": "Not Found", "mb": "Not Found", "kb": "Not Found"}, "ping": "Not Found"})
	} else {
		defer res.Body.Close()
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
		json.Unmarshal(bodyBytes, &data)
		c.JSON(200, data)
	}
}
