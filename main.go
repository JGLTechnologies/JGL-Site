package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Nebulizer1213/GinRateLimit"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"time"
)

var mc *memcache.Client

func getMW(rate int, limit int) func(c *gin.Context) {
	return GinRateLimit.RateLimiter(func(c *gin.Context) string {
		return c.ClientIP() + c.FullPath()
	}, func(c *gin.Context) {
		c.String(429, "Too many requests")
	}, GinRateLimit.InMemoryStore(rate, limit))
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := multitemplate.NewRenderer()
	r.AddFromFiles("home", "go web files/home.html", "go web files/base.html")
	r.AddFromFiles("contact", "go web files/contact.html", "go web files/base.html")
	r.AddFromFiles("404", "go web files/404.html")
	r.AddFromFiles("bmi-home", "go web files/bmi/index.html")
	r.AddFromFiles("bmi-calc", "go web files/bmi/bmi.html")
	r.AddFromFiles("bmi-invalid", "go web files/bmi/invalid.html")
	r.AddFromFiles("contact-thank-you", "go web files/thank-you.html")
	r.AddFromFiles("contact-limit", "go web files/limit.html")
	r.AddFromFiles("contact-captcha", "go web files/captcha.html")
	r.AddFromFiles("contact-bl", "go web files/bl.html")
	r.AddFromFiles("contact-error", "go web files/error.html")
	mc = memcache.New("localhost:8000")
	server := gin.Default()
	server.SetTrustedProxies([]string{"192.168.1.252"})
	server.HTMLRender = r
	server.GET("/", home)
	server.GET("/contact", contact)
	server.GET("/bot", func(c *gin.Context) {
		c.String(200, "JGL Bot documentation is coming soon.")
	})
	test := server.Group("/test")
	{
		test.GET("/bmi", bmiHome)
		test.GET("/bmi/calc", bmiCalc)

	}
	api := server.Group("/api")
	{
		apiMW := getMW(1, 5)
		api.GET("/bot/status", apiMW, botStatus)
		api.GET("/bot/info", apiMW, botInfo)
		api.GET("/dpys", apiMW, dpys)
		api.GET("/aiohttplimiter", apiMW, aiohttpRateLimiter)
		api.POST("/contact", getMW(1, 1), apiContact)

	}
	server.NoRoute(noRoute)
	server.NoMethod(noRoute)
	server.Run(":81")
}

func home(c *gin.Context) {
	c.HTML(200, "home", gin.H{})
}

func contact(c *gin.Context) {
	c.HTML(200, "contact", gin.H{})
}

func bmiHome(c *gin.Context) {
	lastBMI, err := c.Cookie("BMI_LAST")
	if err != nil {
		c.HTML(200, "bmi-home", gin.H{"last": "Not Found"})
	} else {
		c.HTML(200, "bmi-home", gin.H{"last": lastBMI})
	}
}

func bmiCalc(c *gin.Context) {
	var context gin.H
	feet := c.Query("heightft")
	inches := c.Query("heightin")
	weight := c.Query("weight")

	if inches == "" {
		inches = "0"
	}

	feetNum, err := strconv.ParseFloat(feet, 64)
	if err != nil {
		c.HTML(400, "bmi-invalid", gin.H{})
	}

	inchesNum, err := strconv.ParseFloat(inches, 64)
	if err != nil {
		c.HTML(400, "bmi-invalid", gin.H{})
	}

	weightNum, err := strconv.ParseFloat(weight, 64)
	if err != nil {
		c.HTML(400, "bmi-invalid", gin.H{})
	}

	bmi := weightNum / math.Pow((feetNum*12)+inchesNum, 2) * 703

	if bmi > 24.9 {
		newWeight := 24.9 / 703 * math.Pow((feetNum*12)+inchesNum, 2)
		pounds := fmt.Sprintf("%f", math.Round(weightNum-newWeight))
		poundsNum, _ := strconv.Atoi(pounds)
		if poundsNum >= 1 {
			context = gin.H{"bmi": math.Round(bmi), "weight": "You need to loose " + pounds + "pounds to be healthy."}
		} else {
			context = gin.H{"bmi": math.Round(bmi), "weight": ""}
		}
	} else if bmi < 18.5 {
		newWeight := 18.5 / 703 * math.Pow((feetNum*12)+inchesNum, 2)
		pounds := fmt.Sprintf("%f", math.Round(newWeight-weightNum))
		poundsNum, _ := strconv.Atoi(pounds)
		if poundsNum >= 1 {
			context = gin.H{"bmi": math.Round(bmi), "weight": "You need to gain " + pounds + "pounds to be healthy."}
		} else {
			context = gin.H{"bmi": math.Round(bmi), "weight": ""}
		}
	} else {
		context = gin.H{"bmi": math.Round(bmi), "weight": ""}
	}
	maxAge := time.Now().Unix() - time.Date(2038, 1, 1, 0, 0, 0, 0, time.Local).Unix()
	c.SetCookie("BMI_LAST", fmt.Sprintf("%f", bmi), int(maxAge), "/test/bmi", "jgltechnologies.com", true, false)
	c.HTML(200, "bmi-calc", context)
}

func botStatus(c *gin.Context) {
	_, err := http.Get("https://jglbotapi.us/status")
	if err != nil {
		c.JSON(200, gin.H{"online": false})
	} else {
		c.JSON(200, gin.H{"online": true})
	}
}

func botInfo(c *gin.Context) {
	var data map[string]interface{}
	res, err := http.Get("https://jglbotapi.us/info")
	if err != nil {
		c.JSON(200, gin.H{"guilds": "Not Found", "cogs": "Not Found", "shards": "Not Found", "size": gin.H{"gb": "Not Found", "mb": "Not Found", "kb": "Not Found"}, "ping": "Not Found"})
	} else {
		defer res.Body.Close()
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			c.JSON(500, gin.H{"error": err})
		}
		json.Unmarshal(bodyBytes, &data)
		c.JSON(200, data)
	}
}

func dpys(c *gin.Context) {
	var data map[string]map[string]string
	var fileBytes []byte
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
		cached, err := mc.Get("dpys_" + version)
		if err != nil {
			res, err := http.Get("https://raw.githubusercontent.com/Nebulizer1213/dpys/main/dist/dpys-" + version + ".tar.gz")
			if err != nil {
				c.JSON(500, gin.H{"error": err})
				return
			}
			defer res.Body.Close()
			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				c.JSON(500, gin.H{"error": err})
				return
			}
			mc.Set(&memcache.Item{
				Key:        "dpys_" + version,
				Value:      data,
				Flags:      0,
				Expiration: 3600,
			})
			fileBytes = data
		} else {
			fileBytes = cached.Value
		}
		c.JSON(200, gin.H{"version": version, "file_bytes": string(fileBytes)})
	}
}

func aiohttpRateLimiter(c *gin.Context) {
	var data map[string]map[string]string
	var fileBytes []byte
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
		cached, err := mc.Get("aiohttplimiter_" + version)
		if err != nil {
			res, err := http.Get("https://raw.githubusercontent.com/Nebulizer1213/aiohttp-ratelimiter/main/dist/aiohttp-ratelimiter-" + version + ".tar.gz")
			if err != nil {
				c.JSON(500, gin.H{"error": err})
				return
			}
			defer res.Body.Close()
			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				c.JSON(500, gin.H{"error": err})
				return
			}
			mc.Set(&memcache.Item{
				Key:        "aiohttplimiter_" + version,
				Value:      data,
				Flags:      0,
				Expiration: 3600,
			})
			fileBytes = data
		} else {
			fileBytes = cached.Value
		}
		c.JSON(200, gin.H{"version": version, "file_bytes": string(fileBytes)})
	}
}

func apiContact(c *gin.Context) {
	name, exists := c.GetPostForm("name")
	if !exists {
		c.JSON(400, gin.H{"error": "no name was specified"})
		return
	}
	email, exists := c.GetPostForm("email")
	if !exists {
		c.JSON(400, gin.H{"error": "no email was specified"})
		return
	}
	message, exists := c.GetPostForm("message")
	if !exists {
		c.JSON(400, gin.H{"error": "no message was specified"})
		return
	}
	token, exists := c.GetPostForm("token")
	if !exists {
		c.JSON(400, gin.H{"error": "no token was specified"})
		return
	}
	data := map[string]string{"name": name, "email": email, "message": message, "token": token, "ip": c.ClientIP()}
	jsonData, _ := json.Marshal(data)
	res, err := http.Post("https://jglbotapi.us/contact", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		c.HTML(500, "contact-error", gin.H{"error": err})
	} else {
		defer res.Body.Close()
		var resJSON map[string]interface{}
		resData, _ := ioutil.ReadAll(res.Body)
		json.Unmarshal(resData, &resJSON)
		if res.StatusCode == 200 {
			c.HTML(200, "contact-thank-you", gin.H{})
		} else if res.StatusCode == 429 {
			fmt.Println(data)
			c.HTML(429, "contact-limit", gin.H{"remaining": resJSON["remaining"]})
		} else if res.StatusCode == 401 {
			c.HTML(401, "contact-captcha", gin.H{})
		} else if res.StatusCode == 403 {
			c.HTML(403, "contact-bl", gin.H{})
		} else {
			c.HTML(500, "contact-error", gin.H{"error": resJSON["error"]})
		}
	}

}

func noRoute(c *gin.Context) {
	c.HTML(404, "404", gin.H{})
}
