package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

func BotStatus(c *gin.Context) {
	_, err := http.Get("https://jglbotapi.us/status")
	if err != nil {
		c.JSON(200, gin.H{"online": false})
	} else {
		c.JSON(200, gin.H{"online": true})
	}
}

func BotInfo(c *gin.Context) {
	var data map[string]interface{}
	res, err := http.Get("https://jglbotapi.us/info")
	if err != nil {
		c.JSON(200, gin.H{"guilds": "Not Found", "cogs": "Not Found", "shards": "Not Found", "size": gin.H{"gb": "Not Found", "mb": "Not Found", "kb": "Not Found"}, "ping": "Not Found"})
	} else {
		defer res.Body.Close()
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("error: %s", err)})
		}
		json.Unmarshal(bodyBytes, &data)
		c.JSON(200, data)
	}
}
