package web

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/recoilme/dogenews/model"
	. "github.com/stevelacy/daz"
	"github.com/tidwall/interval"
	"github.com/wesleym/telegramwidget"
	"gorm.io/gorm"
)

type Server struct {
	DB   *gorm.DB
	Iv   interval.Interval
	IvEv interval.Interval
	Tg   string
	Usr  *model.User
	Evs  *model.EventBuf
}

// design: https://tailblocks.cc/
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	path = strings.TrimSuffix(path, "/")
	switch r.Method {
	case http.MethodGet:
		switch {
		case path == "ok":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(http.StatusText(200)))
		case path == "px":
			params, err := url.ParseQuery(r.URL.RawQuery)
			if err != nil {
				fmt.Println("Error parse params", err)
				return
			}
			//fmt.Println(params, err)
			ev := parseEvent(params)
			if ev.UserId > 0 {
				s.Evs.Mu.Lock()
				s.Evs.Buf = append(s.Evs.Buf, ev)
				s.Evs.Mu.Unlock()
			}
			w.Header().Set("Content-Type", "image/png")
			w.WriteHeader(http.StatusNoContent)
		case path == "" || path == "td" || path == "ytd" || path == "wk":
			c, err := r.Cookie("usr")
			if err == nil {
				//fmt.Println("cookie", c.Value)
				id, _ := strconv.ParseInt(c.Value, 10, 64)
				usr := &model.User{ID: uint(id)}
				tx := s.DB.Where(usr).First(usr)
				if tx.Error == nil {
					fmt.Printf("usr: %+v\n", usr)
					s.Usr = usr
				}
			}
			bin, err := s.Main(path)
			if checkErr(err, w) {
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(bin)
		case strings.HasPrefix(path, "del"):
			usr := &model.User{TgId: int64(1263310)}
			tx := s.DB.Where(usr).Delete(usr)
			if checkErr(tx.Error, w) {
				return
			}
			w.WriteHeader(http.StatusOK)
		case strings.HasPrefix(path, "imp/"):
			path = strings.TrimPrefix(path, "imp/")
			err := s.Import(path, false)
			if checkErr(err, w) {
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(http.StatusText(200)))

		case strings.HasPrefix(path, "web/"):
			f, err := os.Open(path)
			if err != nil {
				return
			}
			defer f.Close()
			b, _ := ioutil.ReadAll(f)
			// Get the content
			contentType := http.DetectContentType(b[:512])
			if strings.HasSuffix(path, ".svg") {
				contentType = "image/svg+xml"
			}
			//fmt.Println(contentType)
			w.Header().Set("Content-Type", contentType)
			io.Copy(w, bytes.NewReader(b))
		case strings.HasPrefix(path, "auth"):
			params, err := url.ParseQuery(r.URL.RawQuery)
			if checkErr(err, w) {
				return
			}
			u, err := telegramwidget.ConvertAndVerifyForm(params, s.Tg)
			if checkErr(err, w) {
				return
			}
			usr := model.User{TgId: u.ID, AuthDate: time.Now(), Username: u.Username,
				FirstName: u.FirstName, LastName: u.LastName, PhotoURL: fmt.Sprintf("%s", u.PhotoURL)}
			//fmt.Printf("%+v\n", usr)
			res := s.DB.Create(&usr)
			if checkErr(res.Error, w) {
				return
			}
			cookie := http.Cookie{
				Name:    "usr",
				Domain:  "doge.news",
				Value:   fmt.Sprintf("%d", usr.ID),
				Path:    "/",
				Expires: time.Now().Add(365 * 24 * time.Hour),
			}
			http.SetCookie(w, &cookie)
			http.Redirect(w, r, "https://doge.news", http.StatusTemporaryRedirect)
			return
		case path == "rd":
			params, err := url.ParseQuery(r.URL.RawQuery)
			if checkErr(err, w) {
				return
			}
			red := params.Get("urls")
			if red == "" {
				red = "http://" + strings.ToLower(params.Get("url"))
			} else {
				red = "https://" + strings.ToLower(red)
			}
			ev := parseEvent(params)
			if ev.UserId > 0 {
				s.Evs.Mu.Lock()
				s.Evs.Buf = append(s.Evs.Buf, ev)
				s.Evs.Mu.Unlock()
			}
			http.Redirect(w, r, red, http.StatusSeeOther)
		default:
			fmt.Println("def", path)
			w.WriteHeader(http.StatusNotFound)
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func checkErr(err error, w http.ResponseWriter) bool {
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return true
	}
	return false
}

func (s *Server) Main(path string) ([]byte, error) {
	to := time.Now()
	from := to
	switch path {
	case "", "td":
		from = to.Add(time.Duration(-24*1) * time.Hour)
	case "ytd":
		to = to.Add(time.Duration(-24*1) * time.Hour)
		from = to.Add(time.Duration(-24*1) * time.Hour)
	case "wk":
		from = to.Add(time.Duration(-24*7) * time.Hour)
	}

	usrID := uint64(0)
	if s.Usr != nil {
		usrID = uint64(s.Usr.ID)
	}

	art, err := s.ArticlesByDateC(from, to)
	if err != nil {
		return nil, err
	}
	body := H(
		"body", Attr{"class": "text-gray-400 bg-gray-900 body-font"},
		H("div", UnsafeContent(s.Menu())),
		H("div", Attr{"class": "max-w-4xl mx-auto"}, UnsafeContent(Arts(art, path, usrID))),
	)
	html := H("html", head("doge · news"), body)
	return []byte(html()), nil
}

func (s *Server) Menu() string {
	if s.Usr != nil && s.Usr.PhotoURL != "" {
		ava := fmt.Sprintf(`<img class="w-10 h-10 text-white bg-green-500 rounded-full" viewBox="0 0 24 24" src="%s"/>`, s.Usr.PhotoURL)
		return fmt.Sprintf(menu, ava)
	}
	script := `<script async src="https://telegram.org/js/telegram-widget.js?15" data-telegram-login="newsdogebot" data-size="medium" data-radius="4" data-auth-url="https://doge.news/auth" data-request-access="write"></script>`
	return fmt.Sprintf(menu, script)
}

func Arts(art []model.Article, path string, usrID uint64) string {

	items := []string{}
	items = append(items, artHead)

	for i := range art {
		// не статзначимо, занизим метрики
		minimal := 1000
		if art[i].CntView < minimal {
			share := float64(art[i].CntView) / float64(minimal)
			art[i].CntComm = int(float64(art[i].CntComm) * share)
			art[i].CntLike = int(float64(art[i].CntLike) * share)
			continue
		}
		//convert 2 like/comment per miles
		if art[i].CntComm > 0 {
			art[i].CntComm = int(math.Round((float64(art[i].CntComm) * (1000. / float64(art[i].CntView)))))
		}
		if art[i].CntLike > 0 {
			art[i].CntLike = int(math.Round((float64(art[i].CntLike) * (1000. / float64(art[i].CntView)))))
		}

	}

	if path == "" { //main page, score for pubdate
		sort.Slice(art, func(i, j int) bool {
			return art[i].DatePub.Unix() > art[j].DatePub.Unix()
		})
		for i := range art {
			pv, pc := paretto(i, len(art))
			art[i].Score += pv
			art[i].ScoreTxt += pc
		}
	}
	//score for like
	sort.Slice(art, func(i, j int) bool {
		return art[i].CntLike > art[j].CntLike
	})
	for i := range art {
		pv, pc := paretto(i, len(art))
		art[i].Score += pv
		art[i].ScoreTxt += pc
	}
	//score for comments
	sort.Slice(art, func(i, j int) bool {
		return art[i].CntComm > art[j].CntComm
	})
	for i := range art {
		pv, pc := paretto(i, len(art))
		art[i].Score += pv
		art[i].ScoreTxt += pc
	}
	//main page, sort by pubdate inside same score
	if path == "" {
		sort.Slice(art, func(i, j int) bool {
			if int64(art[i].Score*100) == int64(art[j].Score*100) {
				return art[i].DatePub.Unix() > art[j].DatePub.Unix()
			}
			return art[i].Score > art[j].Score
		})
	}
	// not main page, sort by score
	if path != "" {
		sort.Slice(art, func(i, j int) bool {
			return int(art[i].Score*1000) > int(art[j].Score*1000)
		})
	}

	for i, a := range art {
		px := fmt.Sprintf(`<img loading="lazy" width="1" height="2" src="px?r=%d&uid=%d&aid=%d&ev=rndr">`, time.Now().UnixNano(), usrID, a.ID)
		rdurl := ""
		if strings.HasPrefix(a.Url, "https://") {
			rdurl = fmt.Sprintf("urls=%s", strings.ToUpper(strings.TrimPrefix(a.Url, "https://")))
		}
		if strings.HasPrefix(a.Url, "http://") {
			rdurl = fmt.Sprintf("url=%s", strings.ToUpper(strings.TrimPrefix(a.Url, "http://")))
		}
		rd := fmt.Sprintf(`rd?%s&r=%d&uid=%d&aid=%d&ev=clck`, rdurl, time.Now().UnixNano(), usrID, a.ID)
		if i <= 1 || (i > 5 && i < 12) || (i > 21 && i < 42) {
			hero := fmt.Sprintf(artHero2, strings.ToUpper(a.Category), a.TitleMl, a.SummaryMl, rd, px,
				a.AuthorName, fmt.Sprintf("%d", a.CntComm), fmt.Sprintf("%d", a.CntLike))

			items = append(items, hero)
			continue
		}
		element := fmt.Sprintf(artBody, strings.ToUpper(a.Category), a.TitleMl, a.SummaryMl, rd, px,
			fmt.Sprintf("%s", a.ScoreTxt), fmt.Sprintf("%d", a.CntComm), fmt.Sprintf("%d", a.CntLike),
			a.AuthorAva, a.AuthorName, strings.ToUpper(a.Host))
		items = append(items, element)
	}
	if len(items) > 300 {
		items = items[:300] //limit by 300 articles
	}

	items = append(items, artFoot)
	return strings.Join(items, "")
}

func paretto(i, len int) (float64, string) {
	if len == 0 {
		return 0., "D"
	}
	score := float64(i+1) / float64(len)
	switch {
	case score <= 0.2:
		//(.01 - score/100.) - хитрая хрень чтоб в третьем+ знаке сохранить
		//место артикля в оверол выдаче по домену, транслированное в вес
		//0 0.809
		//1 0.808
		//2 0.157
		return 0.80 + (.01 - score/100.), "A"
	case score <= 0.5:
		return 0.15 + (.01 - score/100.), "B"
	case score <= 0.95:
		return 0.04 + (.01 - score/100.), "C"
	default:
		return 0.01 + (.01 - score/100.), "D"
	}
}

func (s *Server) ArticlesByDateC(from, to time.Time) ([]model.Article, error) {
	art := make([]model.Article, 0)
	tx := s.DB.Where("created_at BETWEEN ? AND ?", from, to).Find(&art)
	return art, tx.Error
}

func head(title string) HTML {
	links := H("link", Attr{
		"href": "https://unpkg.com/tailwindcss@^2/dist/tailwind.min.css",
		"rel":  "stylesheet",
	})

	meta := []HTML{
		H("meta", Attr{"charset": "UTF-8"}),
		H("meta", Attr{
			"name":    "viewport",
			"content": "width=device-width, initial-scale=1.0",
		}),
	}

	head := H("head", H("title", title), meta, links)
	return head
}

func parseEvent(params url.Values) *model.Event {
	ev := &model.Event{}
	ev.Event = params.Get("ev")
	uid, _ := strconv.ParseInt(params.Get("uid"), 10, 64)
	ev.UserId = uint(uid)
	aid, _ := strconv.ParseInt(params.Get("aid"), 10, 64)
	ev.ArticleId = uint(aid)
	return ev
}
