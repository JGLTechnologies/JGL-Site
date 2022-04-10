package api

import (
	"github.com/gin-gonic/gin"
	"github.com/imroc/req"
	"time"
)

func BotStatus(c *gin.Context) {
	r := req.New()
	r.SetTimeout(time.Second)
	_, err := r.Get("http://localhost:85/status")
	if err != nil {
		c.AbortWithStatusJSON(200, gin.H{"online": false})
	} else {
		c.JSON(200, gin.H{"online": true})
	}
}

func BotInfo(c *gin.Context) {
	r := req.New()
	r.SetTimeout(time.Second)
	var data map[string]interface{}
	res, err := r.Get("http://localhost:85/info")
	if err != nil {
		c.AbortWithStatusJSON(200, gin.H{"guilds": "Not Found", "cogs": "Not Found", "shards": "Not Found", "size": gin.H{"gb": "Not Found", "mb": "Not Found", "kb": "Not Found"}, "ping": "Not Found"})
	} else {
		err := res.ToJSON(&data)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, data)
	}
}
