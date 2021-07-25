package model

import (
	"sync"
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
	ID        uint      `gorm:"primarykey" json:"uid"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`

	TgId      *int64    `gorm:"index:,unique" json:"tgid,omitempty"`
	AuthDate  time.Time `json:"-"`
	FirstName string    `json:"-"`
	LastName  string    `json:"-"`
	PhotoURL  string    `json:"photo,omitempty"`
	Username  string    `json:"-"`
	Theme     string    `json:"theme,omitempty"`
}

type Event struct {
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"index"`
	UpdatedAt time.Time

	SessionId uint `gorm:"index"`
	Pos       uint
	Event     string `gorm:"index"`
	UserId    uint   `gorm:"index"`
	ArticleId uint
}

type EventBuf struct {
	Mu  sync.Mutex
	Buf []*Event
}
