package models

import "time"

type Not struct {
	Id      string    `gorm:"primaryKey" json:"id"`
	Anime   string    `json:"anime"`
	Episode string    `json:"episode"`
	Date    time.Time `json:"date"`
	Done    bool      `json:"done"`
}
