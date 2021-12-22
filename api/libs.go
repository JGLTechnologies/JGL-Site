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

func Versions(c *gin.Context) {
	data := make(map[string]string)
	var grl map[string]string
	var pmrl map[string]string
	var dpys map[string]map[string]string
	var aiohttplimiter map[string]map[string]string

	client := http.Client{
		Timeout: time.Second * 5,
	}

	res, grlErr := client.Get("https://api.github.com/repos/Nebulizer1213/GinRateLimit/releases/latest")
	if grlErr != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": grlErr})
	} else {
		defer res.Body.Close()
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("%s", err)})
			return
		}
		json.Unmarshal(bodyBytes, &grl)
		version := grl["name"]
		data["GinRateLimit"] = version
	}

	res, pmrlErr := client.Get("https://api.github.com/repos/Nebulizer1213/precise-memory-rate-limit/releases/latest")
	if pmrlErr != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": pmrlErr})
	} else {
		defer res.Body.Close()
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("%s", err)})
			return
		}
		json.Unmarshal(bodyBytes, &pmrl)
		version := pmrl["name"]
		data["precise-memory-rate-limit"] = version
	}

	res, dpysErr := client.Get("https://pypi.org/pypi/dpys/json")
	if dpysErr != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": dpysErr})
	} else {
		defer res.Body.Close()
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("%s", err)})
			return
		}
		json.Unmarshal(bodyBytes, &dpys)
		version := dpys["info"]["version"]
		data["dpys"] = version
	}

	res, aiohttplimiterErr := client.Get("https://pypi.org/pypi/aiohttp-ratelimiter/json")
	if aiohttplimiterErr != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": aiohttplimiterErr})
	} else {
		defer res.Body.Close()
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": fmt.Sprintf("%s", err)})
			return
		}
		json.Unmarshal(bodyBytes, &aiohttplimiter)
		version := aiohttplimiter["info"]["version"]
		data["aiohttp-ratelimiter"] = version
	}

	c.JSON(200, data)

}
