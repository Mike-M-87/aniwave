package notifications

import (
	"aniwave/models"
	"aniwave/structure"
	"aniwave/utils"
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

	duplicateKeyError = `ERROR: duplicate key value violates unique constraint "nots_pkey" (SQLSTATE 23505)`
)

func initAniClient() {
	aniClient = &http.Client{}
	myCookie = &http.Cookie{
		Name:       "session",
		Value:      "HJhP2ILDSQkIL0BIM0iCAP8EjWzRFnCrfYvGpp7e",
		Domain:     "aniwave.to",
		Path:       "/",
		RawExpires: "2023-11-21T12:03:40.351Z",
	}
}

func FetchAllNotifications() {
	var page int = 1
	for i := page; i <= page; i++ {
		nots, err := GetNotifications(i)
		if err != nil || len(nots) <= 0 {
			continue
		}

		for _, v := range nots {
			err = utils.DB.Create(v).Error
			if err != nil && err.Error() == duplicateKeyError {
				return
			}
		}
		page++
		time.Sleep(time.Millisecond * 100)
	}
}

func GetNotifications(currentPage int) ([]*models.Not, error) {
	initAniClient()
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
		if v.Name == "session" {
			myCookie = v
		}
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
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
			println(err)
			continue
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
	unitMap := map[string]time.Duration{
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
	duration, exists := unitMap[unit]
	if !exists {
		return time.Time{}, fmt.Errorf("unsupported time unit")
	}
	currentTime := time.Now()
	// Calculate the absolute time by subtracting the duration from the current time
	absoluteTime := currentTime.Add(-time.Duration(value) * duration)
	fmt.Println(absoluteTime)
	return absoluteTime, nil
}