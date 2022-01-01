package utils

import (
	"fmt"
	"github.com/JGLTechnologies/GinRateLimit"
	"github.com/gin-gonic/gin"
	"github.com/imroc/req"
	"github.com/mattn/go-isatty"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func StartsWith(s string, sw string) bool {
	swLen := len(sw)
	sLen := len(s)
	if swLen > sLen {
		return false
	} else if s[:swLen] == sw {
		return true
	} else {
		return false
	}
}

func GetPythonLibDownloads(project string) string {
	var data map[string]interface{}
	client := http.Client{
		Timeout: time.Second * 5,
	}
	request := req.New()
	request.SetClient(&client)
	res, err := request.Get("https://api.pepy.tech/api/projects/" + project)
	if err != nil || res.Response().StatusCode != 200 {
		return "Not Found"
	}
	jsonErr := res.ToJSON(&data)
	if jsonErr != nil {
		return "Not Found"
	}
	return strconv.Itoa(int(data["total_downloads"].(float64)))
}

func GetNPMLibDownloads(project string) string {
	var date string
	date += strconv.Itoa(time.Now().Year())
	date += strconv.Itoa(int(time.Now().Month()))
	date += strconv.Itoa(time.Now().Day())
	var data map[string]interface{}
	client := http.Client{
		Timeout: time.Second * 5,
	}
	request := req.New()
	request.SetClient(&client)
	res, err := request.Get("https://api.npmjs.org/downloads/point/2020-1-1:" + date + "/" + project)
	if err != nil || res.Response().StatusCode != 200 {
		return "Not Found"
	}
	jsonErr := res.ToJSON(&data)
	if jsonErr != nil {
		return "Not Found"
	}
	return strconv.Itoa(int(data["downloads"].(float64)))
}

func GetGoLibDownloads(project string) string {
	request := req.New()
	var data map[string]interface{}
	client := http.Client{
		Timeout: time.Second * 5,
	}
	request.SetClient(&client)
	header := make(http.Header)
	header.Set("Authorization", "token "+os.Getenv("gh_token"))
	res, err := request.Get("https://api.github.com/repos/JGLTechnologies/"+project+"/traffic/clones?per=week", header)
	if err != nil || res.Response().StatusCode != 200 {
		return "Not Found"
	}
	jsonErr := res.ToJSON(&data)
	if jsonErr != nil {
		return "Not Found"
	}
	return strconv.Itoa(int(data["uniques"].(float64)))
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
