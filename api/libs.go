package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

func DPYS(c *gin.Context) {
	var data map[string]map[string]string
	res, err := http.Get("https://pypi.org/pypi/dpys/json")
	if err != nil {
		c.JSON(500, gin.H{"error": err})
	} else {
		defer res.Body.Close()
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			c.JSON(500, gin.H{"error": err})
			return
		}
		json.Unmarshal(bodyBytes, &data)
		version := data["info"]["version"]
		c.JSON(200, gin.H{"version": version})
	}
}

func AIOHTTPRateLimiter(c *gin.Context) {
	var data map[string]map[string]string
	res, err := http.Get("https://pypi.org/pypi/aiohttp-ratelimiter/json")
	if err != nil {
		c.JSON(500, gin.H{"error": err})
	} else {
		defer res.Body.Close()
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			c.JSON(500, gin.H{"error": err})
			return
		}
		json.Unmarshal(bodyBytes, &data)
		version := data["info"]["version"]
		c.JSON(200, gin.H{"version": version})
	}
}
