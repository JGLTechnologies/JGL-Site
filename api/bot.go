package api

import (
	"github.com/gin-gonic/gin"
	"github.com/imroc/req/v3"
	"time"
)

const botServiceURL = "http://localhost:85"

var botClient = req.C().SetTimeout(time.Second)

var botInfoNotFound = gin.H{
	"guilds":   "Not Found",
	"cogs":     "Not Found",
	"shards":   "Not Found",
	"size":     gin.H{"gb": "Not Found", "mb": "Not Found", "kb": "Not Found"},
	"ping":     "Not Found",
	"tickets":  "Not Found",
	"messages": "Not Found",
}

func BotStatus(c *gin.Context) {
	_, err := botClient.R().Get(botServiceURL + "/status")
	if err != nil {
		c.AbortWithStatusJSON(200, gin.H{"online": false})
		return
	}

	c.JSON(200, gin.H{"online": true})
}

func BotInfo(c *gin.Context) {
	var data map[string]interface{}

	res, err := botClient.R().Get(botServiceURL + "/info")
	if err != nil {
		c.AbortWithStatusJSON(200, botInfoNotFound)
		return
	}

	err = res.UnmarshalJson(&data)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, data)
}
