package main

import (
	"JGLSite/api"
	"JGLSite/test"
	"JGLSite/utils"
	"encoding/json"
	"fmt"
	"github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/imroc/req"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var Store *persist.MemoryStore

func main() {
	godotenv.Load("../.env")
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
	server := gin.New()
	server.HTMLRender = r
	Store = persist.NewMemoryStore(time.Hour)

	server.Use(gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		c.HTML(500, "contact-error", gin.H{"error": fmt.Sprintf("%s", err)})
		c.AbortWithStatus(500)
	}))
	server.Use(utils.LoggerWithConfig(gin.LoggerConfig{}))
	server.SetTrustedProxies([]string{"192.168.1.252", "127.0.0.1", "192.168.1.1"})

	server.GET("/", cache.CacheByRequestPath(Store, time.Hour*24), home)
	server.GET("/home", cache.CacheByRequestPath(Store, time.Hour*24), home)
	server.GET("/contact", cache.CacheByRequestPath(Store, time.Hour*24), contact)
	server.GET("/logo.png", cache.CacheByRequestPath(Store, time.Hour*24), logo)
	server.GET("/favicon.ico", cache.CacheByRequestPath(Store, time.Hour*24), favicon)

	testGroup := server.Group("/test")
	{
		testGroup.GET("/bmi", test.BMIHome)
		testGroup.GET("/bmi/calc", test.BMICalc)
	}

	apiGroup := server.Group("/api")
	{
		apiGroup.GET("/bot/status", cache.CacheByRequestPath(Store, time.Minute), api.BotStatus)
		apiGroup.GET("/bot/info", cache.CacheByRequestPath(Store, time.Hour), api.BotInfo)
		apiGroup.GET("/dpys", cache.CacheByRequestPath(Store, time.Minute*10), api.DPYS)
		apiGroup.GET("/aiohttplimiter", cache.CacheByRequestPath(Store, time.Minute*10), api.AIOHTTPRateLimiter)
		apiGroup.GET("/GinRateLimit", cache.CacheByRequestPath(Store, time.Minute*10), api.GinRateLimit)
		apiGroup.GET("/precise-memory-rate-limit", cache.CacheByRequestPath(Store, time.Minute*10), api.PreciseMemoryRateLimit)
		apiGroup.GET("/versions", versions)
		apiGroup.GET("/downloads", downloads)
		apiGroup.POST("/contact", utils.GetMW(1, 1), api.Contact)
	}

	server.NoRoute(noRoute)
	server.NoMethod(noRoute)
	log.Fatal(server.Run(":81"))
}

func downloads(c *gin.Context) {
	c.JSON(200, gin.H{
		"dpys":                      utils.GetPythonLibDownloads("dpys", Store),
		"aiohttp-ratelimiter":       utils.GetPythonLibDownloads("aiohttp-ratelimiter", Store),
		"precise-memory-rate-limit": utils.GetNPMLibDownloads("precise-memory-rate-limit", Store),
		"GinRateLimit":              utils.GetGoLibDownloads("GinRateLimit", Store),
	})
}

func favicon(c *gin.Context) {
	c.File("static/favicon.ico")
}

func logo(c *gin.Context) {
	c.File("static/logo.png")
}

func home(c *gin.Context) {
	c.HTML(200, "home", gin.H{})
}

func contact(c *gin.Context) {
	c.HTML(200, "contact", gin.H{})
}

func noRoute(c *gin.Context) {
	c.HTML(404, "404", gin.H{})
}

func versions(c *gin.Context) {
	var grlValue string
	var pmrlValue string
	var dpysValue string
	var aiohttplimiterValue string
	data := make(map[string]string)
	var grl map[string]string
	var pmrl map[string]string
	var dpys map[string]map[string]string
	var aiohttplimiter map[string]map[string]string

	client := http.Client{
		Timeout: time.Second * 5,
	}

	header := make(http.Header)
	header.Set("Authorization", "token "+os.Getenv("gh_token"))
	request := req.New()
	request.SetClient(&client)

	if grlErr := Store.Get("grl_version", &grlValue); grlErr != nil {
		res, resErr := request.Get("https://api.github.com/repos/Nebulizer1213/GinRateLimit/releases/latest", header)
		if resErr != nil || res.Response().StatusCode != 200 {
			Store.Set("grl_version", "Not Found", time.Minute*10)
			data["GinRateLimit"] = "Not Found"
		} else {
			err := res.ToJSON(&grl)
			if err != nil {
				Store.Set("grl_version", "Not Found", time.Minute*10)
				data["GinRateLimit"] = "Not Found"
			}
			version := grl["name"]
			data["GinRateLimit"] = version
			Store.Set("grl_version", version, time.Minute*10)
		}
	} else {
		data["GinRateLimit"] = grlValue
	}

	if pmrlErr := Store.Get("pmrl_version", &pmrlValue); pmrlErr != nil {
		res, resErr := request.Get("https://api.github.com/repos/Nebulizer1213/precise-memory-rate-limit/releases/latest", header)
		if resErr != nil || res.Response().StatusCode != 200 {
			Store.Set("pmrl_version", "Not Found", time.Minute*10)
			data["precise-memory-rate-limit"] = "Not Found"
		} else {
			err := res.ToJSON(&pmrl)
			if err != nil {
				Store.Set("pmrl_version", "Not Found", time.Minute*10)
				data["precise-memory-rate-limit"] = "Not Found"
			}
			version := pmrl["name"]
			data["precise-memory-rate-limit"] = version
			Store.Set("pmrl_version", version, time.Minute*10)
		}
	} else {
		data["precise-memory-rate-limit"] = pmrlValue
	}

	if dpysErr := Store.Get("dpys_version", &dpysValue); dpysErr != nil {
		res, resErr := client.Get("https://pypi.org/pypi/dpys/json")
		if resErr != nil || res.StatusCode != 200 {
			Store.Set("dpys_version", "Not Found", time.Minute*10)
			data["dpys"] = "Not Found"
		} else {
			defer res.Body.Close()
			bodyBytes, err := ioutil.ReadAll(res.Body)
			if err != nil {
				Store.Set("dpys_version", "Not Found", time.Minute*10)
				data["dpys"] = "Not Found"
			}
			json.Unmarshal(bodyBytes, &dpys)
			version := "v" + dpys["info"]["version"]
			data["dpys"] = version
			Store.Set("dpys_version", version, time.Minute*10)
		}
	} else {
		data["dpys"] = dpysValue
	}

	if aiohttplimiterErr := Store.Get("aiohttplimiter_version", &aiohttplimiterValue); aiohttplimiterErr != nil {
		res, resErr := client.Get("https://pypi.org/pypi/aiohttp-ratelimiter/json")
		if resErr != nil || res.StatusCode != 200 {
			Store.Set("aiohttplimiter_version", "Not Found", time.Minute*10)
			data["aiohttp-ratelimiter"] = "Not Found"
		} else {
			defer res.Body.Close()
			bodyBytes, err := ioutil.ReadAll(res.Body)
			if err != nil {
				Store.Set("aiohttplimiter_version", "Not Found", time.Minute*10)
				data["aiohttp-ratelimiter"] = "Not Found"
			}
			json.Unmarshal(bodyBytes, &aiohttplimiter)
			version := "v" + aiohttplimiter["info"]["version"]
			data["aiohttp-ratelimiter"] = version
			Store.Set("aiohttplimiter_version", version, time.Minute*10)
		}
	} else {
		data["aiohttp-ratelimiter"] = aiohttplimiterValue
	}

	c.JSON(200, data)

}
