package api

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Announcement struct {
	Title  string
	Body   string
	Time   int64
	Expire int64
}

type JNAData struct {
	Announcements []Announcement `json:"announcements"`
	Emails        []string       `json:"emails"`
}

var jnaFileMu sync.Mutex

func ReadJNAData() (JNAData, error) {
	jnaFileMu.Lock()
	defer jnaFileMu.Unlock()
	return readJNAData()
}

func WriteJNAData(data JNAData) error {
	jnaFileMu.Lock()
	defer jnaFileMu.Unlock()

	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("jna.json", encoded, 0600)
}

func readJNAData() (JNAData, error) {
	var data JNAData
	contents, err := os.ReadFile("jna.json")
	if os.IsNotExist(err) || len(contents) == 0 {
		return data, nil
	}
	if err != nil {
		return data, err
	}
	if err := json.Unmarshal(contents, &data); err != nil {
		return JNAData{}, err
	}
	return data, nil
}

func JNA(c *gin.Context) {
	data, err := ReadJNAData()
	if err != nil {
		c.String(http.StatusInternalServerError, "Unable to read announcements")
		return
	}
	finalAnnouncements := []Announcement{}
	for _, announcement := range data.Announcements {
		if announcement.Expire > time.Now().Unix() {
			finalAnnouncements = append(finalAnnouncements, announcement)
		}
	}
	c.JSON(http.StatusOK, finalAnnouncements)
}
