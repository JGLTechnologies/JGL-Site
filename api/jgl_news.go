package api

import (
	"JGLSite/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func JNA(c *gin.Context) {
	var a []utils.Announcement
	utils.DB.Model(&utils.Announcement{}).Where("expire > ?", time.Now().Unix()).Order("ID DESC").Find(&a)
	c.JSON(http.StatusOK, a)
}
