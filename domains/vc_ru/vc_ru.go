package vc_ru

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

type Vc struct {
	HostName string
	bow      *browser.Browser
}

func New(hostName string) *Vc {
	br := surf.NewBrowser()
	br.SetUserAgent("Mozilla/5.0 (Windows NT 6.3; x64; rv:31.0) Gecko/20100101 Firefox/31.0")
	return &Vc{HostName: hostName, bow: br}
}

func (vc *Vc) Host() string {
	return vc.HostName
}

func (vc *Vc) Links() ([]string, error) {
	err := vc.bow.Open("https://vc.ru/new")
	if err != nil {
		return nil, err
	}
	links := make([]string, 0, 10)
	vc.bow.Find("a.content-link").Each(func(_ int, s *goquery.Selection) {
		if link, ok := s.Attr("href"); ok {
			if l, ok := vc.LinkOk(link, false); ok {
				links = append(links, l)
			}
		}
	})
	return links, nil
}

func (vc *Vc) LinkOk(link string, check bool) (string, bool) {
	if !strings.Contains(link, vc.HostName) {
		return "", false
	}
	ind := strings.Index(link, "-")
	res := link
	if ind > 0 {
		res = link[:ind]
		indLast := strings.LastIndex(res, "/")
		if indLast > 0 {
			indFirst := strings.LastIndex(res[:indLast], "/")
			if indFirst > 0 {
				res = res[:indFirst] + res[indLast:]
			}
		}
		//TODO: convert to https://vc.ru/245468 ?
	}
	if !check {
		return res, true
	}
	err := vc.bow.Head(link)
	if err != nil {
		return "", false
	}
	if vc.bow.StatusCode() == 403 || vc.bow.StatusCode() == 404 {
		return "", false
	}
	if vc.bow.StatusCode() != 200 {
		fmt.Println(vc.bow.StatusCode(), link)
	}
	if vc.bow.StatusCode() == 429 {
		time.Sleep(30 * time.Second)
	}
	return res, true
}

func (vc *Vc) Article(link string) (*model.Article, error) {
	err := vc.bow.Open(link)
	if err != nil {
		return nil, err
	}
	a := &model.Article{}
	a.StatusCode = vc.bow.StatusCode()
	a.Title = strings.TrimSpace(vc.bow.Find("h1").First().Text())
	width := ""
	height := ""
	vc.bow.Find("meta").Each(func(index int, item *goquery.Selection) {

		switch item.AttrOr("property", "") {
		case "og:title":
			if a.Title == "" {
				a.Title = strings.TrimSpace(item.AttrOr("content", ""))
			}
			a.Title = strings.TrimSpace(strings.Replace(a.Title, "Статьи редакции", "", -1))
		case "og:description":
			a.Summary = strings.TrimSpace(item.AttrOr("content", ""))
		case "og:image":
			a.ImageBanner = strings.TrimSpace(item.AttrOr("content", ""))
		case "og:image:width":
			width = strings.TrimSpace(item.AttrOr("content", ""))
		case "og:image:height":
			height = strings.TrimSpace(item.AttrOr("content", ""))
		case "article:published_time":
			if datePub, err := time.Parse(time.RFC3339, item.AttrOr("content", "")); err == nil {
				a.DatePub = datePub
			}
		case "article:author":
			if a.AuthorName == "" {
				a.AuthorName = strings.TrimSpace(item.AttrOr("content", ""))
			}
		case "article:publisher":
			a.AuthorUrl = strings.TrimSpace(item.AttrOr("content", ""))
		}
		if item.AttrOr("name", "") == "author" {
			a.AuthorName = strings.TrimSpace(item.AttrOr("content", ""))
			//fmt.Println(a.AuthorName)
		}
	})
	if width != "" && height != "" {
		a.ImageBannerMeta = fmt.Sprintf("width=%s&height=%s", width, height)
	}
	a.Language = "ru"
	a.Category = strings.ToLower(strings.TrimSpace(vc.bow.Find("div.content-header-author__name").First().Text()))

	cntViewS := strings.TrimSpace(vc.bow.Find("span.views__value").First().Text())
	if cntView, err := strconv.ParseInt(cntViewS, 10, 64); err == nil {
		a.CntView = int(cntView)
	}

	dc := vc.bow.Find("div.vote--content").First()
	cntLikeS := dc.AttrOr("data-count", "")
	if cntLike, err := strconv.ParseInt(cntLikeS, 10, 64); err == nil {
		a.CntLike = int(cntLike)
	}

	cntComS := strings.TrimSpace(vc.bow.Find("span.comments_counter__count__value").First().Text())
	if cntCom, err := strconv.ParseInt(cntComS, 10, 64); err == nil {
		a.CntComm = int(cntCom)
	}

	sel := vc.bow.Find("div.content-header-author__avatar").First()
	if sel != nil {
		if val, ok := sel.Attr("style"); ok {
			val = strings.TrimPrefix(val, "background-image: url('")
			val = strings.TrimSuffix(val, "')")
			a.AuthorAva = val
		}
	}
	cnt := vc.bow.Find("div.content--full").First()
	if cnt != nil {
		cnt.Find(".l-hidden").Remove()
		cnt.Find(".content-info").Remove()
		if html, err := cnt.Html(); err == nil {
			a.ContentHtml = html
		}
		a.ContentText = strings.Join(strings.Fields(cnt.Text()), " ")
	}
	return a, nil
}
