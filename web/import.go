package web

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/joshleeb/textrank"
	"github.com/recoilme/dogenews/domains/habr_com"
	"github.com/recoilme/dogenews/domains/vc_ru"
	"github.com/recoilme/dogenews/model"
)

func (s *Server) Import(path string, validate bool) error {
	fmt.Println("Import", path, time.Now(), " ", validate)
	time.Sleep(1 * time.Second)
	var site model.Site
	switch path {
	case "vc_ru":
		site = vc_ru.New("vc.ru")
	case "habr_com":
		site = habr_com.New("habr.com")
	default:
		return errors.New("site not found:" + path)
	}

	// import
	links, err := site.Links()
	if err != nil {
		return err
	}
	if validate {
		art, err := s.ArticlesByDateC(time.Now().Add(-24*time.Hour), time.Now(), nil)
		if err != nil {
			fmt.Println(err)
		}
		for _, a := range art {
			links = append(links, a.Url)
		}
	}
	ins, upd, del := 0, 0, 0
	for _, link := range links {
		host, err := hostName(link)
		if host != site.Host() {
			//fmt.Println("host != site.Host()", link, host, site.Host()) //TODO skip this?
			continue
		}
		time.Sleep(1 * time.Second)
		if validate {
			time.Sleep(5 * time.Second)
		}

		if err != nil {
			fmt.Println(link, host, err) //TODO skip this?
			continue
		}

		art := &model.Article{Host: host, Url: link}
		//find in db by host/url
		s.DB.Where(art).First(art)

		//import article
		a, err := site.Article(link)
		if err != nil {
			fmt.Println(link, err)
			continue
		}
		a.Host = art.Host
		a.Url = art.Url
		if art.ID == 0 && a.StatusCode == 200 {

			sentences := textrank.RankSentences(a.Title+". "+a.Summary+" "+a.ContentText, 5)
			for _, sent := range sentences {
				sent := strings.TrimSpace(sent)
				ll := len([]rune(sent))
				if ll > 20 && ll < 200 && a.TitleMl == "" && sent != a.SummaryMl {
					a.TitleMl = sent
					if len(a.Title) < len(a.TitleMl) {
						a.TitleMl = a.Title
						a.SummaryMl = sent
						break
					}
					continue
				}
				if ll > 20 && ll < 300 && a.SummaryMl == "" && sent != a.TitleMl {
					a.SummaryMl = sent
					continue
				}
			}
			res := s.DB.Create(a)
			if res.Error != nil {
				return res.Error
			}
			ins++
			continue
		}
		//article present..
		_, ok := site.LinkOk(link, true)
		if !ok {
			// and deleted on site (403,404 http code)
			//s.DB.Delete(art.Url)
			s.DB.Delete(&model.Article{}, art.ID)
			fmt.Println("delete", link)
			del++
			continue
		}
		// update article
		art.CntComm = a.CntComm
		art.CntLike = a.CntLike
		art.CntShare = a.CntShare
		art.CntView = a.CntView
		//a.ID = art.ID
		res := s.DB.Save(art)
		if res.Error != nil {
			return res.Error
		}
		upd++

	}
	fmt.Printf("ins:%d upd:%d del:%d\n", ins, upd, del)
	return nil
}

// hostName return hostName without www.
func hostName(uri string) (domain string, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		return
	}
	domain = strings.ToLower(strings.Replace(u.Hostname(), "www.", "", 1))
	return domain, nil
}
