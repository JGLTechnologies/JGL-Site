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

func Contact(c *gin.Context) {
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
	data := map[string]string{"name": name, "email": email, "message": message, "token": token, "ip": utils.GetIP(c)}
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
