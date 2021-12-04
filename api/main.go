package api

import (
	"JGLSite/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

type postForm struct {
	name    string `form:"name" binding:"required"`
	email   string `form:"email" binding:"required"`
	message string `form:"message" binding:"required"`
	token   string `form:"token" binding:"required"`
}

func Contact(c *gin.Context) {
	formData := postForm{}
	if err := c.Bind(formData); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}
	name := formData.name
	email := formData.email
	message := formData.message
	token := formData.token
	data := map[string]string{"name": name, "email": email, "message": message, "token": token, "ip": utils.GetIP(c)}
	jsonData, _ := json.Marshal(data)
	res, err := http.Post("https://jglbotapi.us/contact", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		c.HTML(500, "contact-error", gin.H{"error": fmt.Sprintf("error: %s", err)})
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
