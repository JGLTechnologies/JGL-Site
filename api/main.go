package api

import (
	"JGLSite/utils"
	"fmt"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/imroc/req/v3"
	"log"
	"math"
	"os"
	"time"
)

var client = req.C().SetTimeout(time.Second * 5)

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
	id := requestid.Get(c)
	formData := postForm{}
	if bindingErr := c.ShouldBind(&formData); bindingErr != nil {
		c.HTML(400, "client-error", gin.H{"message": "The request body you provided is invalid.", "title": "Invalid request body"})
		return
	}
	name := formData.Name
	email := formData.Email
	message := formData.Message
	token := formData.Token
	if len(name) > 200 || len(email) > 254 || len(message) > 1020 {
		c.HTML(400, "client-error", gin.H{"message": "The form body you provided is invalid.", "title": "Invalid form body"})
		return
	}
	data := map[string]string{"name": name, "email": email, "message": message, "token": token, "ip": c.ClientIP(), "ua": c.GetHeader("User-Agent")}
	res, err := client.R().SetBodyJsonMarshal(&data).Post("http://localhost:85/contact")
	if err != nil {
		errStruct := &utils.Err{Message: err.Error(), Date: time.Now().Format("Jan 02, 2006 3:04:05 pm"), ID: id, IP: c.ClientIP(), Path: c.Request.URL.String()}
		utils.Pool.Submit(func() {
			utils.DB.Create(errStruct)
		})
		c.HTML(500, "error", gin.H{"id": id})
		c.AbortWithStatus(500)
	} else {
		var resJSON interface{}
		jsonErr := res.UnmarshalJson(&resJSON)
		if jsonErr != nil {
			errStruct := &utils.Err{Message: jsonErr.Error(), Date: time.Now().Format("Jan 02, 2006 3:04:05 pm"), ID: id, IP: c.ClientIP(), Path: c.Request.URL.String()}
			utils.Pool.Submit(func() {
				utils.DB.Create(errStruct)
			})
			c.HTML(500, "error", gin.H{"id": id})
			c.AbortWithStatus(500)
		} else {
			if res.IsSuccessState() {
				c.HTML(200, "contact-thank-you", gin.H{})
			} else if res.StatusCode == 429 {
				minutes := math.Trunc(time.Duration(resJSON.(map[string]interface{})["remaining"].(float64) * float64(time.Minute.Nanoseconds())).Minutes())
				seconds := math.Trunc(time.Duration(resJSON.(map[string]interface{})["remaining"].(float64) * float64(time.Second.Nanoseconds())).Seconds())
				var remaining string
				if minutes < 1 {
					remaining = fmt.Sprintf("%v seconds", seconds)
				} else {
					remaining = fmt.Sprintf("%v minutes", minutes)
				}
				c.HTML(429, "contact-limit", gin.H{"remaining": remaining})
			} else if res.StatusCode == 401 {
				c.HTML(401, "contact-captcha", gin.H{})
			} else if res.StatusCode == 403 {
				if res.String() == "blacklist" {
					c.HTML(403, "contact-bl", gin.H{})
				} else {
					c.HTML(403, "contact-spam", gin.H{})
				}
			} else {
				errStruct := &utils.Err{Message: resJSON.(map[string]interface{})["error"].(string), Date: time.Now().Format("Jan 02, 2006 3:04:05 pm"), ID: id, IP: c.ClientIP(), Path: c.Request.URL.String()}
				utils.Pool.Submit(func() {
					utils.DB.Create(errStruct)
				})
				c.HTML(500, "error", gin.H{"id": id})
				c.AbortWithStatus(500)
			}
		}
	}
}

func GetKSPData(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	if c.GetHeader("Key") != os.Getenv("KSP_API") {
		c.String(403, "Invalid Code")
		return
	}

	// --- Step 1: fetch msToken ---
	tenantID := os.Getenv("AZURE_TENANT_ID")
	clientID := os.Getenv("AZURE_CLIENT_ID")
	clientSecret := os.Getenv("AZURE_CLIENT_SECRET")

	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", tenantID)

	tokenResp, err := client.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"grant_type":    "client_credentials",
			"client_id":     clientID,
			"client_secret": clientSecret,
			"resource":      "https://manage.devcenter.microsoft.com",
		}).
		Post(tokenURL)
	if err != nil {
		log.Printf("Token request failed: %v", err)
		c.String(500, "Failed to get token")
		return
	}

	var tokenData struct {
		AccessToken string `json:"access_token"`
	}
	if err := tokenResp.UnmarshalJson(&tokenData); err != nil {
		log.Printf("Failed to parse token response: %v", err)
		c.String(500, "Invalid token response")
		return
	}

	msToken := tokenData.AccessToken

	fmt.Println(msToken)

	// --- Step 2: query Partner Center with msToken ---
	appId := os.Getenv("KSP_ID")
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")

	url := fmt.Sprintf(
		"https://manage.devcenter.microsoft.com/v1.0/my/analytics/acquisitions?applicationId=%s&aggregationLevel=day&startDate=%s&endDate=%s",
		appId, startDate, endDate,
	)

	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+msToken).
		SetHeader("Content-Type", "application/json").
		Get(url)
	if err != nil {
		log.Printf("Data request failed: %v", err)
		c.String(500, "Failed to fetch app data")
		return
	}

	c.String(200, resp.String())
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
