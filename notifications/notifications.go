package notifications

import (
	"aniwave/models"
	"aniwave/structure"
	"aniwave/utils"
	"os"

	// "bytes"
	// "encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	myCookie  *http.Cookie
	aniClient *http.Client
	unitMap   = map[string]time.Duration{
		"week":    7 * 24 * time.Hour,
		"weeks":   7 * 24 * time.Hour,
		"day":     24 * time.Hour,
		"days":    24 * time.Hour,
		"hour":    time.Hour,
		"hours":   time.Hour,
		"minute":  time.Minute,
		"minutes": time.Minute,
		"second":  time.Second,
		"seconds": time.Second,
	}
)

const duplicateKeyError = `ERROR: duplicate key value violates unique constraint "nots_pkey" (SQLSTATE 23505)`

func initAniClient() {
	if aniClient == nil || myCookie == nil {
		aniClient = &http.Client{}
		myCookie = &http.Cookie{
			Name:       "session",
			Value:      os.Getenv("ANIWAVE_COOKIE"),
			Domain:     "aniwave.to",
			Path:       "/",
			RawExpires: "2023-11-22T16:03:40.351Z",
		}
	}
}

func FetchAllNotifications() {
	initAniClient()
	var page int = 1
	for i := page; i <= page; i++ {
		nots, err := GetNotifications(i)
		if err != nil || len(nots) <= 0 {
			continue
		}
		fmt.Println(len(nots))
		for _, v := range nots {
			err = utils.DB.Create(v).Error
			if err == nil {
				// SendTelegramNotification(v)
				fmt.Println(utils.AsPrettyJson(v))
				time.Sleep(time.Second)
			} else if err != nil && err.Error() == duplicateKeyError {
				return
			}
		}
		page++
		time.Sleep(time.Second)
	}
}

func GetNotifications(currentPage int) ([]*models.Not, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://aniwave.to/user/notification?page=%d", currentPage), nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(myCookie)
	resp, err := aniClient.Do(req)
	if err != nil {
		return nil, err
	}
	for _, v := range resp.Cookies() {
		if v != nil && v.Name == "session" {
			myCookie = v
			fmt.Println(v.Value)
		}
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	WriteToFile(string(responseBody))
	reId := regexp.MustCompile(`<a data-id="(\d+)"`)
	notIds := reId.FindAllString(string(responseBody), -1)

	reInfo := regexp.MustCompile(`<div class="info">[\s\S]*?</time>[\s\S]*?</div>`)
	infoDivs := reInfo.FindAllString(string(responseBody), -1)

	nots := make([]*models.Not, len(infoDivs))

	for i, v := range infoDivs {
		var infoDivXml structure.DivInfo
		err = xml.Unmarshal([]byte(v), &infoDivXml)
		if err != nil {
			continue
		}
		content := infoDivXml.Div.Span
		date, err := parseRelativeTime(infoDivXml.Time)
		if err != nil {
			fmt.Println("parse time err", date)
		}
		nots[i] = &models.Not{
			Id:      extractNotID(notIds[i]),
			Anime:   content[0].Text,
			Episode: content[1].Text,
			Date:    date,
		}
	}
	return nots, nil
}

func WriteToFile(s string) {
	f, err := os.OpenFile("resp.html", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return
	}
	_, err = f.WriteString(s)
	if err != nil {
		return
	}
}

func extractNotID(s string) string {
	re := regexp.MustCompile(`(\d+)`)
	match := re.FindString(s)
	return match
}

func parseRelativeTime(relativeTime string) (time.Time, error) {
	// Split the relative time string to extract the value and unit
	parts := strings.Fields(relativeTime)
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("invalid relative time format")
	}
	value, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid relative time value")
	}
	unit := parts[1]
	// Map relative time units to their respective durations

	duration, exists := unitMap[unit]
	if !exists {
		return time.Time{}, fmt.Errorf("unsupported time unit")
	}
	var currentTime time.Time
	location, err := time.LoadLocation("Africa/Nairobi")
	if err != nil {
		fmt.Println("Error loading location:", err)
		currentTime = time.Now()
	} else {
		currentTime = time.Now().In(location)
	}
	absoluteTime := currentTime.Add(-time.Duration(value) * duration)
	return absoluteTime, nil
}

func SendTelegramNotification(not *models.Not) error {
	initAniClient()
	payload := strings.NewReader(fmt.Sprintf("{\n\t\"text\":\"\\n‼️New Episode Out‼️\\n\\n------------------\\n\\nAnime:   %s\\n\\nEpisode: %s\\n\\nDate:      %s\\n\\nWatch:   [Link](https://aniwave.to/user/notification/read/%s)\\n\\n------------------\\n\\n✨✨\"\n}", not.Anime, not.Episode, not.Date.Format(time.ANSIC), not.Id))
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=-1001738363628&parse_mode=markdown", os.Getenv("TELBOT_KEY")), payload)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	_, err = aniClient.Do(req)
	if err != nil {
		return err
	}
	return nil
}
