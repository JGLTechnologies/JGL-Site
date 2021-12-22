package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"time"
)

func DPYS(c *gin.Context) {
	var data map[string]map[string]string
	client := http.Client{
		Timeout: time.Second * 5,
	}
	res, err := client.Get("https://pypi.org/pypi/dpys/json")
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("%s", err)})
	} else {
		defer res.Body.Close()
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("%s", err)})
			return
		}
		json.Unmarshal(bodyBytes, &data)
		version := data["info"]["version"]
		c.JSON(200, gin.H{"version": version})
	}
}

func GinRateLimit(c *gin.Context) {
	var data map[string]interface{}
	client := http.Client{
		Timeout: time.Second * 5,
	}
	res, err := client.Get("https://api.github.com/repos/Nebulizer1213/GinRateLimit/releases/latest")
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("%s", err)})
	} else {
		defer res.Body.Close()
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("%s", err)})
			return
		}
		json.Unmarshal(bodyBytes, &data)
		version := data["name"]
		fmt.Println(data)
		c.JSON(200, gin.H{"version": version})
	}
}

func PreciseMemoryRateLimit(c *gin.Context) {
	var data map[string]interface{}
	client := http.Client{
		Timeout: time.Second * 5,
	}
	res, err := client.Get("https://api.github.com/repos/Nebulizer1213/precise-memory-rate-limit/releases/latest")
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("%s", err)})
	} else {
		defer res.Body.Close()
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("%s", err)})
			return
		}
		json.Unmarshal(bodyBytes, &data)
		version := data["name"]
		fmt.Println(data)
		c.JSON(200, gin.H{"version": version})
	}
}

func AIOHTTPRateLimiter(c *gin.Context) {
	var data map[string]map[string]string
	client := http.Client{
		Timeout: time.Second * 5,
	}
	res, err := client.Get("https://pypi.org/pypi/aiohttp-ratelimiter/json")
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err})
	} else {
		defer res.Body.Close()
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("%s", err)})
			return
		}
		json.Unmarshal(bodyBytes, &data)
		version := data["info"]["version"]
		c.JSON(200, gin.H{"version": version})
	}
}
