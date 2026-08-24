package api

import (
	"JGLSite/utils"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/imroc/req/v3"
)

var client = req.C().SetTimeout(time.Second * 5)

const cfGraphQL = "https://api.cloudflare.com/client/v4/graphql"

type postForm struct {
	Name    string `form:"name" binding:"required"`
	Email   string `form:"email" binding:"required"`
	Message string `form:"message" binding:"required"`
	Token   string `form:"token" binding:"required"`
}

type Project struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
}

func Contact(c *gin.Context) {
	var formData postForm
	if err := c.ShouldBind(&formData); err != nil {
		utils.RenderErrorPage(c, http.StatusBadRequest, "The request body you provided is invalid.")
		return
	}
	if len(formData.Name) > 200 || len(formData.Email) > 254 || len(formData.Message) > 1020 {
		utils.RenderErrorPage(c, http.StatusBadRequest, "The form body you provided is invalid.")
		return
	}
	data := map[string]string{
		"name":    formData.Name,
		"email":   formData.Email,
		"message": formData.Message,
		"token":   formData.Token,
		"ip":      c.ClientIP(),
		"ua":      c.GetHeader("User-Agent"),
	}
	res, err := client.R().SetBodyJsonMarshal(&data).Post("http://localhost:85/contact")
	if err != nil {
		panic(err)
	}

	var resJSON interface{}
	if err := res.UnmarshalJson(&resJSON); err != nil {
		panic(err)
	}

	if res.IsSuccessState() {
		c.HTML(200, "contact-thank-you", gin.H{})
		return
	}

	if res.StatusCode == http.StatusTooManyRequests {
		utils.RenderErrorPage(c, http.StatusTooManyRequests, fmt.Sprintf(
			"Too many form submissions. Please try again in %s.",
			formatRemaining(resJSON),
		))
		return
	}

	if res.StatusCode == http.StatusUnauthorized {
		utils.RenderErrorPage(c, http.StatusUnauthorized,
			"Our security check could not verify your request. Please return to the contact form and try again.")
		return
	}

	if res.StatusCode == http.StatusForbidden {
		if res.String() == "blacklist" {
			utils.RenderErrorPage(c, http.StatusForbidden,
				"You have been blacklisted from submitting forms. If you believe this is a mistake, contact support.")
		} else {
			utils.RenderErrorPage(c, http.StatusForbidden,
				"Our automated filters flagged the submission as possible spam. Please review your message and try again.")
		}
		return
	}

	if response, ok := resJSON.(map[string]interface{}); ok {
		if message, ok := response["error"].(string); ok && strings.TrimSpace(message) != "" {
			panic(errors.New(message))
		}
	}
	panic(fmt.Errorf("contact service returned status %d", res.StatusCode))
}

func CFProxy(c *gin.Context) {
	cfToken := os.Getenv("cfToken")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusBadRequest, "read body: %v", err)
		return
	}
	resp, err := client.R().
		SetHeaders(map[string]string{
			"Authorization": "Bearer " + cfToken,
			"Content-Type":  "application/json",
		}).
		SetBodyBytes(body).
		Post(cfGraphQL)
	if err != nil {
		c.String(http.StatusBadGateway, "upstream error: %v", err)
		return
	}

	hopByHop := map[string]struct{}{
		"Connection":          {},
		"Keep-Alive":          {},
		"Proxy-Authenticate":  {},
		"Proxy-Authorization": {},
		"TE":                  {},
		"Trailers":            {},
		"Transfer-Encoding":   {},
		"Upgrade":             {},
	}
	for k, vv := range resp.Header {
		if _, skip := hopByHop[http.CanonicalHeaderKey(k)]; skip {
			continue
		}
		for _, v := range vv {
			c.Writer.Header().Add(k, v)
		}
	}
	c.Status(resp.StatusCode)
	c.String(resp.StatusCode, strings.TrimSpace(resp.String()))
}

func GetErr(c *gin.Context) {
	query := struct {
		ID string `form:"id" binding:"required"`
	}{}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid id"})
		return
	}
	_, uuidErr := uuid.Parse(query.ID)
	if uuidErr != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "invalid id"})
		return
	}
	err := &utils.Err{}
	res := utils.DB.First(err, "id=?", query.ID)
	if res.RowsAffected < 1 {
		c.AbortWithStatusJSON(400, gin.H{"error": "there is no error with that id"})
	} else {
		c.JSON(200, err)
	}
}

func formatRemaining(resJSON interface{}) string {
	response, ok := resJSON.(map[string]interface{})
	if !ok {
		return "a short while"
	}
	remaining, ok := response["remaining"].(float64)
	if !ok || remaining <= 0 || math.IsNaN(remaining) || math.IsInf(remaining, 0) {
		return "a short while"
	}

	minutes := math.Trunc(time.Duration(remaining * float64(time.Minute)).Minutes())
	seconds := math.Trunc(time.Duration(remaining * float64(time.Second)).Seconds())

	if minutes < 1 {
		return fmt.Sprintf("%v seconds", seconds)
	}

	return fmt.Sprintf("%v minutes", minutes)
}
