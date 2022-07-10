package api

import (
	"JGLSite/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/imroc/req/v3"
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

type botForm struct {
	Name        string `form:"name" binding:"required"`
	Email       string `form:"email" binding:"required"`
	Description string `form:"desc" binding:"required"`
	Token       string `form:"token" binding:"required"`
}

type Project struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Downloads   string `json:"downloads"`
	Private     bool   `json:"private"`
}

func Contact(c *gin.Context) {
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
	data := map[string]string{"name": name, "email": email, "message": message, "token": token, "ip": c.ClientIP()}
	res, err := client.R().SetBodyJsonMarshal(&data).Post("http://localhost:85/contact")
	if err != nil {
		id, _ := uuid.NewRandom()
		errStruct := &utils.Err{Message: err.Error(), Date: time.Now().Format("Jan 02, 2006 3:04:05 pm"), ID: id.String()}
		utils.Pool.Submit(func() {
			utils.DB.Create(errStruct)
		})
		c.HTML(500, "error", gin.H{"id": id.String()})
		c.AbortWithStatus(500)
	} else {
		var resJSON interface{}
		jsonErr := res.UnmarshalJson(&resJSON)
		if jsonErr != nil {
			id, _ := uuid.NewRandom()
			errStruct := &utils.Err{Message: jsonErr.Error(), Date: time.Now().Format("Jan 02, 2006 3:04:05 pm"), ID: id.String()}
			utils.Pool.Submit(func() {
				utils.DB.Create(errStruct)
			})
			c.HTML(500, "error", gin.H{"id": id.String()})
			c.AbortWithStatus(500)
		} else {
			if res.IsSuccess() {
				c.HTML(200, "contact-thank-you", gin.H{})
			} else if res.StatusCode == 429 {
				c.HTML(429, "contact-limit", gin.H{"remaining": resJSON.(map[string]interface{})["remaining"]})
			} else if res.StatusCode == 401 {
				c.HTML(401, "contact-captcha", gin.H{})
			} else if res.StatusCode == 403 {
				c.HTML(403, "contact-bl", gin.H{})
			} else {
				id, _ := uuid.NewRandom()
				errStruct := &utils.Err{Message: resJSON.(map[string]interface{})["error"].(string), Date: time.Now().Format("Jan 02, 2006 3:04:05 pm"), ID: id.String()}
				utils.Pool.Submit(func() {
					utils.DB.Create(errStruct)
				})
				c.HTML(500, "error", gin.H{"id": id.String()})
				c.AbortWithStatus(500)
			}
		}
	}
}

func CustomBot(c *gin.Context) {
	formData := botForm{}
	if bindingErr := c.ShouldBind(&formData); bindingErr != nil {
		c.HTML(400, "client-error", gin.H{"message": "The request body you provided is invalid.", "title": "Invalid request body"})
		return
	}
	name := formData.Name
	email := formData.Email
	desc := formData.Description
	token := formData.Token
	if len(name) > 200 || len(email) > 254 || len(desc) > 1020 {
		c.HTML(400, "client-error", gin.H{"message": "The form body you provided is invalid.", "title": "Invalid form body"})
		return
	}
	data := map[string]string{"name": name, "email": email, "desc": desc, "token": token, "ip": c.ClientIP()}
	res, err := client.R().SetBodyJsonMarshal(&data).Post("http://localhost:85/custom-bot")
	if err != nil {
		id, _ := uuid.NewRandom()
		errStruct := &utils.Err{Message: err.Error(), Date: time.Now().Format("Jan 02, 2006 3:04:05 pm"), ID: id.String()}
		utils.Pool.Submit(func() {
			utils.DB.Create(errStruct)
		})
		c.HTML(500, "error", gin.H{"id": id.String()})
		c.AbortWithStatus(500)
	} else {
		var resJSON interface{}
		jsonErr := res.UnmarshalJson(&resJSON)
		if jsonErr != nil {
			id, _ := uuid.NewRandom()
			errStruct := &utils.Err{Message: jsonErr.Error(), Date: time.Now().Format("Jan 02, 2006 3:04:05 pm"), ID: id.String()}
			utils.Pool.Submit(func() {
				utils.DB.Create(errStruct)
			})
			c.HTML(500, "error", gin.H{"id": id.String()})
			c.AbortWithStatus(500)
		} else {
			if res.IsSuccess() {
				c.HTML(200, "contact-thank-you", gin.H{})
			} else if res.StatusCode == 429 {
				c.HTML(429, "contact-limit", gin.H{"remaining": resJSON.(map[string]interface{})["remaining"]})
			} else if res.StatusCode == 401 {
				c.HTML(401, "contact-captcha", gin.H{})
			} else if res.StatusCode == 403 {
				c.HTML(403, "contact-bl", gin.H{})
			} else {
				id, _ := uuid.NewRandom()
				errStruct := &utils.Err{Message: resJSON.(map[string]interface{})["error"].(string), Date: time.Now().Format("Jan 02, 2006 3:04:05 pm"), ID: id.String()}
				utils.Pool.Submit(func() {
					utils.DB.Create(errStruct)
				})
				c.HTML(500, "error", gin.H{"id": id.String()})
				c.AbortWithStatus(500)
			}
		}
	}
}

func Projects() ([]*Project, error) {
	dpys := utils.GetPythonLibDownloads("dpys")
	aiohttplimiter := utils.GetPythonLibDownloads("aiohttp-ratelimiter")
	sf := utils.GetGoLibDownloads("SimpleFiles")
	pmrl := utils.GetNPMLibDownloads("precise-memory-rate-limit")
	grl := utils.GetGoLibDownloads("GinRateLimit")
	downloads := map[string]string{
		"DPYS":                      dpys,
		"aiohttp-ratelimiter":       aiohttplimiter,
		"precise-memory-rate-limit": pmrl,
		"GinRateLimit":              grl,
		"SimpleFiles":               sf,
		"total":                     GetTotal([]string{dpys, aiohttplimiter, pmrl, grl}),
	}

	res, err := client.R().SetHeader("Authorization", "token "+os.Getenv("gh_token")).Get("https://api.github.com/orgs/JGLTechnologies/repos")
	if err != nil || res.IsError() {
		return []*Project{}, err
	} else {
		var data []*Project
		jsonErr := res.UnmarshalJson(&data)
		if jsonErr != nil {
			return []*Project{}, jsonErr
		} else {
			for _, v := range data {
				d, ok := downloads[v.Name]
				if ok {
					v.Downloads = d
				} else {
					v.Downloads = ""
				}
			}
			return data, nil
		}
	}
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
	}
	err := &utils.Err{}
	res := utils.DB.First(err, "id=?", query.ID)
	if res.RowsAffected < 1 {
		c.AbortWithStatusJSON(400, gin.H{"error": "there is no error with that id"})
	} else {
		c.JSON(200, err)
	}
}
