package model

import (
	"time"
)

type Site interface {
	Host() string
	Links() ([]string, error)
	LinkOk(link string, check bool) (string, bool)
	Article(link string) (*Article, error)
}

type Article struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	//DeletedAt       gorm.DeletedAt `gorm:"index"`
	Host            string `gorm:"index"`
	Url             string `gorm:"index:,unique"`
	Title           string
	TitleMl         string
	Summary         string
	SummaryMl       string
	ImageMain       string
	ImageMainMeta   string
	ImageBanner     string
	ImageBannerMeta string
	DatePub         time.Time `gorm:"index"`
	DateMod         time.Time
	AuthorName      string
	AuthorUrl       string
	AuthorAva       string
	Language        string `gorm:"index"`
	Category        string `gorm:"index"`
	Tags            string
	CntView         int
	CntLike         int
	CntComm         int
	CntShare        int
	ContentText     string
	ContentHtml     string

	Score      float64
	ScoreTxt   string
	StatusCode int
}

type User struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	TgId      int64 `gorm:"index:,unique"`
	AuthDate  time.Time
	FirstName string
	LastName  string
	PhotoURL  string
	Username  string
}
