package api

import (
	"JGLSite/utils"
	"github.com/gin-gonic/gin"
	"strconv"
)

func GetTotal(list []string) string {
	total := 0
	for _, v := range list {
		if v == "Not Found" {
			continue
		} else {
			num, _ := strconv.Atoi(v)
			total += num
		}
	}
	return strconv.Itoa(total)
}

func Versions(c *gin.Context) {
	data := utils.Versions()
	c.JSON(200, data)
}

func Downloads(c *gin.Context) {
	dpys := utils.GetPythonLibDownloads("dpys")
	aiohttplimiter := utils.GetPythonLibDownloads("aiohttp-ratelimiter")
	pmrl := utils.GetNPMLibDownloads("precise-memory-rate-limit")
	grl := utils.GetGoLibDownloads("GinRateLimit")
	c.JSON(200, gin.H{
		"dpys":                      dpys,
		"aiohttp-ratelimiter":       aiohttplimiter,
		"precise-memory-rate-limit": pmrl,
		"GinRateLimit":              grl,
		"total":                     GetTotal([]string{dpys, aiohttplimiter, pmrl, grl}),
	})
}
