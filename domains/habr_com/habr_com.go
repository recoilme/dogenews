package habr_com

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/headzoo/surf"
	"github.com/headzoo/surf/browser"
	"github.com/recoilme/dogenews/model"
)

type Habr struct {
	HostName string
	bow      *browser.Browser
}

func New(hostName string) *Habr {
	br := surf.NewBrowser()
	br.SetUserAgent("Mozilla/5.0 (Windows NT 6.3; x64; rv:31.0) Gecko/20100101 Firefox/31.0")
	return &Habr{HostName: hostName, bow: br}
}

func (habr *Habr) Host() string {
	return habr.HostName
}

func (habr *Habr) Links() ([]string, error) {
	err := habr.bow.Open("https://habr.com/ru/rss/all/all/?fl=ru")
	if err != nil {
		return nil, err
	}
	links := make([]string, 0, 10)
	habr.bow.Find("guid").Each(func(_ int, s *goquery.Selection) {
		//fmt.Println(s.Text())
		if l, ok := habr.LinkOk(s.Text(), false); ok {
			links = append(links, l)
		}

	})
	return links, nil
}

func (habr *Habr) LinkOk(link string, check bool) (string, bool) {

	if !strings.Contains(link, habr.HostName) {
		return "", false
	}
	res := link

	if !check {
		return res, true
	}
	err := habr.bow.Head(link)
	if err != nil {
		return "", false
	}
	if habr.bow.StatusCode() == 403 || habr.bow.StatusCode() == 404 {
		return "", false
	}
	if habr.bow.StatusCode() != 200 {
		fmt.Println(habr.bow.StatusCode(), link)
	}
	if habr.bow.StatusCode() == 429 {
		time.Sleep(30 * time.Second)
	}
	return res, true
}

func (habr *Habr) Article(link string) (*model.Article, error) {
	a := &model.Article{}

	err := habr.bow.Open(link)
	if err != nil {
		return nil, err
	}
	a.StatusCode = habr.bow.StatusCode()
	a.Title = strings.TrimSpace(habr.bow.Find("h1").First().Text())

	width := ""
	height := ""

	habr.bow.Find("meta").Each(func(index int, item *goquery.Selection) {

		switch item.AttrOr("property", "") {
		case "og:title":
			if a.Title == "" {
				a.Title = strings.TrimSpace(item.AttrOr("content", ""))
			}
		case "og:description":
			a.Summary = strings.TrimSpace(item.AttrOr("content", ""))
		case "og:image":
			if a.ImageBanner == "" {
				a.ImageBanner = strings.TrimSpace(item.AttrOr("content", ""))
			}
		case "og:image:width":
			width = strings.TrimSpace(item.AttrOr("content", ""))
		case "og:image:height":
			height = strings.TrimSpace(item.AttrOr("content", ""))
		case "aiturec:datetime":
			if datePub, err := time.Parse(time.RFC3339, item.AttrOr("content", "")); err == nil {
				a.DatePub = datePub
			}

		}
	})
	if width != "" && height != "" {
		a.ImageBannerMeta = fmt.Sprintf("width=%s&height=%s", width, height)
	}
	a.Language = "ru"
	a.Category = strings.ToLower(strings.TrimSpace(habr.bow.Find("a.hub-link").First().Text()))
	cntViewS := strings.TrimSpace(habr.bow.Find("span.post-stats__views-count").First().Text())
	k := 1
	if strings.HasSuffix(cntViewS, "k") {
		//115k
		k = 1000
		cntViewS = cntViewS[:len(cntViewS)-1]
	}
	if strings.Contains(cntViewS, ",") {
		//3,5k
		cntViewS = strings.Replace(cntViewS, ",", ".", -1)
	}
	if cntView, err := strconv.ParseFloat(cntViewS, 64); err == nil {
		a.CntView = int(cntView * float64(k))
	}

	cntLikeS := habr.bow.Find("span.voting-wjt__counter").First().Text()
	if strings.HasPrefix(cntLikeS, "+") || strings.HasPrefix(cntLikeS, "-") {
		cntLikeS = cntLikeS[1:]
	}
	if cntLike, err := strconv.ParseInt(cntLikeS, 10, 64); err == nil {
		a.CntLike = int(cntLike)
	}

	cntComS := strings.TrimSpace(habr.bow.Find("span.post-stats__comments-count").First().Text())
	if cntCom, err := strconv.ParseInt(cntComS, 10, 64); err == nil {
		a.CntComm = int(cntCom)
	}
	a.AuthorName = habr.bow.Find("span.user-info__nickname").First().Text()
	habr.bow.Find("a.post__user-info").Each(func(_ int, s *goquery.Selection) {
		if link, ok := s.Attr("href"); ok {
			a.AuthorUrl = link

		}
	})
	habr.bow.Find("img.user-info__image-pic").Each(func(_ int, s *goquery.Selection) {
		if link, ok := s.Attr("src"); ok {
			if a.AuthorAva == "" {
				a.AuthorAva = link
			}
		}
	})
	if strings.HasPrefix(a.AuthorAva, "//") {
		a.AuthorAva = "https:" + a.AuthorAva
	}
	cnt := habr.bow.Find("div.post__body_full").First()
	if cnt != nil {
		if html, err := cnt.Html(); err == nil {
			a.ContentHtml = html
		}
		a.ContentText = strings.Join(strings.Fields(cnt.Text()), " ")
	}

	return a, nil
}
