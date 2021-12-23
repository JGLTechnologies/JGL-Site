package utils

import (
	"encoding/json"
	"fmt"
	"github.com/Nebulizer1213/GinRateLimit"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"github.com/imroc/req"
	"github.com/mattn/go-isatty"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func GetPythonLibDownloads(project string, store *persist.MemoryStore) string {
	var downloads string
	if err := store.Get("downloads_"+project, &downloads); err != nil {
		var data map[string]interface{}
		client := http.Client{
			Timeout: time.Second * 5,
		}
		res, err := client.Get("https://api.pepy.tech/api/projects/" + project)
		if err != nil || res.StatusCode != 200 {
			store.Set("downloads_"+project, "Not Found", time.Minute*10)
			return "Not Found"
		}
		defer res.Body.Close()
		bodyBytes, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			store.Set("downloads_"+project, "Not Found", time.Minute*10)
			return "Not Found"
		}
		jsonErr := json.Unmarshal(bodyBytes, &data)
		if jsonErr != nil {
			fmt.Println(fmt.Sprintf("%s", jsonErr))
			store.Set("downloads_"+project, "Not Found", time.Minute*10)
			return "Not Found"
		}
		store.Set("downloads_"+project, strconv.Itoa(int(data["total_downloads"].(float64))), time.Hour*24)
		return strconv.Itoa(int(data["total_downloads"].(float64)))
	} else {
		return downloads
	}
}

func GetNPMLibDownloads(project string, store *persist.MemoryStore) string {
	var date string
	date += strconv.Itoa(time.Now().Year())
	date += strconv.Itoa(int(time.Now().Month()))
	date += strconv.Itoa(time.Now().Day())
	var downloads string
	if err := store.Get("downloads_"+project, &downloads); err != nil {
		var data map[string]interface{}
		client := http.Client{
			Timeout: time.Second * 5,
		}
		res, err := client.Get("https://api.npmjs.org/downloads/point/2020-1-1:" + date + "/" + project)
		if err != nil || res.StatusCode != 200 {
			store.Set("downloads_"+project, "Not Found", time.Minute*10)
			return "Not Found"
		}
		defer res.Body.Close()
		bodyBytes, readErr := ioutil.ReadAll(res.Body)
		if readErr != nil {
			store.Set("downloads_"+project, "Not Found", time.Minute*10)
			return "Not Found"
		}
		jsonErr := json.Unmarshal(bodyBytes, &data)
		if jsonErr != nil {
			fmt.Println(fmt.Sprintf("%s", jsonErr))
			store.Set("downloads_"+project, "Not Found", time.Minute*10)
			return "Not Found"
		}
		store.Set("downloads_"+project, strconv.Itoa(int(data["downloads"].(float64))), time.Hour*24)
		return strconv.Itoa(int(data["downloads"].(float64)))
	} else {
		return downloads
	}
}

func GetGoLibDownloads(project string, store *persist.MemoryStore) string {
	var downloads string
	if err := store.Get("downloads_"+project, &downloads); err != nil {
		request := req.New()
		var data map[string]interface{}
		client := http.Client{
			Timeout: time.Second * 5,
		}
		request.SetClient(&client)
		header := make(http.Header)
		header.Set("Authorization", "token "+os.Getenv("gh_token"))
		res, err := request.Get("https://api.github.com/repos/Nebulizer1213/"+project+"/traffic/clones?per=week", header)
		if err != nil || res.Response().StatusCode != 200 {
			store.Set("downloads_"+project, "Not Found", time.Minute*10)
			return "Not Found"
		}
		jsonErr := res.ToJSON(&data)
		if jsonErr != nil {
			store.Set("downloads_"+project, "Not Found", time.Minute*10)
			return "Not Found"
		}
		store.Set("downloads_"+project, strconv.Itoa(int(data["uniques"].(float64))), time.Hour*24)
		return strconv.Itoa(int(data["uniques"].(float64)))
	} else {
		return downloads
	}
}

func Versions(store *persist.MemoryStore) map[string]string {
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

	res, grlErr := request.Get("https://api.github.com/repos/Nebulizer1213/GinRateLimit/releases/latest", header)
	if grlErr != nil || res.Response().StatusCode != 200 {
		data["GinRateLimit"] = "Not Found"
	} else {
		err := res.ToJSON(&grl)
		if err != nil {
			data["GinRateLimit"] = "Not Found"
		}
		version := grl["name"]
		data["GinRateLimit"] = version
	}

	res, pmrlErr := request.Get("https://api.github.com/repos/Nebulizer1213/precise-memory-rate-limit/releases/latest", header)
	if pmrlErr != nil || res.Response().StatusCode != 200 {
		data["precise-memory-rate-limit"] = "Not Found"
	} else {
		err := res.ToJSON(&pmrl)
		if err != nil {
			data["precise-memory-rate-limit"] = "Not Found"
		}
		version := pmrl["name"]
		data["precise-memory-rate-limit"] = version
	}

	res, dpysErr := request.Get("https://pypi.org/pypi/dpys/json")
	if dpysErr != nil || res.Response().StatusCode != 200 {
		data["dpys"] = "Not Found"
	} else {
		err := res.ToJSON(&dpys)
		if err != nil {
			data["dpys"] = "Not Found"
		}
		version := "v" + dpys["info"]["version"]
		data["dpys"] = version
		store.Set("dpys_version", version, time.Minute*10)
	}
	res, aiohttplimiterErr := request.Get("https://pypi.org/pypi/aiohttp-ratelimiter/json")
	if aiohttplimiterErr != nil || res.Response().StatusCode != 200 {
		data["aiohttp-ratelimiter"] = "Not Found"
	} else {
		err := res.ToJSON(&aiohttplimiter)
		if err != nil {
			store.Set("aiohttplimiter_version", "Not Found", time.Minute*10)
			data["aiohttp-ratelimiter"] = "Not Found"
		}
		version := "v" + aiohttplimiter["info"]["version"]
		data["aiohttp-ratelimiter"] = version
	}
	return data
}

func GetMW(rate int, limit int) func(c *gin.Context) {
	return GinRateLimit.RateLimiter(func(c *gin.Context) string {
		return GetIP(c) + c.FullPath()
	}, func(c *gin.Context) {
		c.String(429, "Too many requests")
	}, GinRateLimit.InMemoryStore(rate, limit))
}

func GetIP(c *gin.Context) string {
	ip := c.GetHeader("X-Forwarded-For")
	if ip == "" {
		ip = c.ClientIP()
	}
	ip = strings.Split(ip, ",")[0]
	return ip
}

func defaultLogFormatter(param gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		// Truncate in a golang < 1.8 safe way
		param.Latency = param.Latency - param.Latency%time.Second
	}
	return fmt.Sprintf("[GIN] %v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
		param.ErrorMessage,
	)
}

func LoggerWithConfig(conf gin.LoggerConfig) gin.HandlerFunc {
	formatter := conf.Formatter
	if formatter == nil {
		formatter = defaultLogFormatter
	}

	out := conf.Output
	if out == nil {
		out = gin.DefaultWriter
	}

	notlogged := conf.SkipPaths

	if w, ok := out.(*os.File); !ok || os.Getenv("TERM") == "dumb" ||
		(!isatty.IsTerminal(w.Fd()) && !isatty.IsCygwinTerminal(w.Fd())) {
	}

	var skip map[string]struct{}

	if length := len(notlogged); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range notlogged {
			skip[path] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log only when path is not being skipped
		if _, ok := skip[path]; !ok {
			param := gin.LogFormatterParams{
				Request: c.Request,
				Keys:    c.Keys,
			}

			// Stop timer
			param.TimeStamp = time.Now()
			param.Latency = param.TimeStamp.Sub(start)

			param.ClientIP = GetIP(c)
			param.Method = c.Request.Method
			param.StatusCode = c.Writer.Status()
			param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()

			param.BodySize = c.Writer.Size()

			if raw != "" {
				path = path + "?" + raw
			}

			param.Path = path

			fmt.Fprint(out, formatter(param))
		}
	}
}
