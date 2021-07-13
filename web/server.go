package web

import (
	"bytes"
	"encoding/json"
	"errors"
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

	"github.com/cristalhq/jwt/v3"
	"github.com/recoilme/dogenews/model"
	"github.com/tidwall/interval"
	"github.com/wesleym/telegramwidget"
	"gorm.io/gorm"
)

type Server struct {
	DB    *gorm.DB
	Iv    interval.Interval
	IvEv  interval.Interval
	Evs   *model.EventBuf
	Token []byte
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
			c, err := r.Cookie("doge")
			usr := &model.User{}
			if err == nil {
				usr, err = s.userCurr(c.Value)
				if checkErr(err, w) {
					return
				}
			} else {
				cookie, err := s.userUps(r.URL.Host, usr)
				if checkErr(err, w) {
					return
				}
				http.SetCookie(w, cookie)
				err = nil
			}
			//fmt.Printf("usr:%+v\n", usr)
			params, err := url.ParseQuery(r.URL.RawQuery)
			if checkErr(err, w) {
				return
			}
			bin, err := s.Main(path, params, usr)
			if checkErr(err, w) {
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(bin)
		case strings.HasPrefix(path, "del"):
			me := int64(1263310)
			usr := &model.User{TgId: &me}
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

		case strings.HasPrefix(path, "m/") || strings.HasPrefix(path, "favicon/"):
			if strings.HasPrefix(path, "m/") {
				path = "web/" + path[2:]
			}
			if strings.HasPrefix(path, "favicon/") {
				path = "web/" + path
			}
			f, err := os.Open(path)
			if err != nil {
				return
			}
			defer f.Close()
			b, _ := ioutil.ReadAll(f)
			// Get the content
			lenb := len(b)
			if lenb > 512 {
				lenb = 512
			}
			contentType := http.DetectContentType(b[:lenb])
			if strings.HasSuffix(path, ".svg") {
				contentType = "image/svg+xml"
			}
			if strings.HasSuffix(path, ".css") {
				contentType = "text/css"
			}
			//fmt.Println(contentType)
			w.Header().Set("Content-Type", contentType)
			io.Copy(w, bytes.NewReader(b))
		case strings.HasPrefix(path, "auth"):
			params, err := url.ParseQuery(r.URL.RawQuery)
			if checkErr(err, w) {
				return
			}

			u, err := telegramwidget.ConvertAndVerifyForm(params, string(s.Token))
			if checkErr(err, w) {
				return
			}
			usr := &model.User{TgId: &u.ID, AuthDate: time.Now(), Username: u.Username,
				FirstName: u.FirstName, LastName: u.LastName, PhotoURL: fmt.Sprintf("%s", u.PhotoURL)}
			cookie, err := s.userUps(r.URL.Host, usr)
			if checkErr(err, w) {
				return
			}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, "https://doge.news", http.StatusTemporaryRedirect)
			return
		case path == "rd":
			params, err := url.ParseQuery(r.URL.RawQuery)
			if checkErr(err, w) {
				return
			}
			red := params.Get("urls")
			if red == "" {
				red = params.Get("urls")
				red = "http://" + strings.ToLower(red)
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
		case path == "st":
			var count int64
			s.DB.Model(&model.User{}).Count(&count)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("%d", count)))
		case path == "api/v1":
			params, err := url.ParseQuery(r.URL.RawQuery)
			if checkErr(err, w) {
				return
			}
			th := params.Get("theme")
			if !(th == "light" || th == "dark") {
				err = errors.New("Wrong theme")
				if checkErr(err, w) {
					return
				}
			}
			c, err := r.Cookie("doge")
			if checkErr(err, w) {
				return
			}
			usr, err := s.userCurr(c.Value)
			if checkErr(err, w) {
				return
			}
			usr.Theme = th
			cookie, err := s.userUps(r.URL.Host, usr)
			if checkErr(err, w) {
				return
			}
			http.SetCookie(w, cookie)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(http.StatusText(200)))
			return
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

func (s *Server) Main(path string, params url.Values, usr *model.User) ([]byte, error) {
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

	usrID := uint64(usr.ID)

	art, err := s.ArticlesByDateC(from, to, params)
	if err != nil {
		return nil, err
	}
	tgLogin := `<script async src="https://telegram.org/js/telegram-widget.js?15" data-telegram-login="newsdogebot" data-size="medium" data-radius="4" data-auth-url="https://doge.news/auth" data-request-access="write"></script>`
	if usr != nil && usr.PhotoURL != "" {
		tgLogin = `<img src="` + usr.PhotoURL + `" style="width:2em; height: 2em; margin-top: -0.5em; border-radius: 50%;">`
	}
	themeIco := "☪"
	theme := "light"
	if usr.Theme == "dark" {
		themeIco = "☀"
		theme = "dark"
	}
	html := fmt.Sprintf(html_, theme, "doge · news", tgLogin, themeIco, Arts(art, path, usrID))

	return []byte(html), nil
}

func Arts(art []model.Article, path string, usrID uint64) string {

	items := []string{}

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

	for _, a := range art {
		if a.StatusCode != 200 {
			continue
		}
		px := fmt.Sprintf(`<img loading="lazy" width="1" height="2" src="px?r=%d&uid=%d&aid=%d&ev=rndr">`, time.Now().UnixNano(), usrID, a.ID)
		rdurl := ""
		if strings.HasPrefix(a.Url, "https://") {
			rdurl = fmt.Sprintf("urls=%s", strings.ToUpper(strings.TrimPrefix(a.Url, "https://")))
		}
		if strings.HasPrefix(a.Url, "http://") {
			rdurl = fmt.Sprintf("url=%s", strings.ToUpper(strings.TrimPrefix(a.Url, "http://")))
		}
		rd := fmt.Sprintf(`rd?%s&r=%d&uid=%d&aid=%d&ev=clck`, rdurl, time.Now().UnixNano(), usrID, a.ID)
		categ := ""
		if a.Category != "" {
			categ = fmt.Sprintf("<a href='?c=%s'>%s</a>", url.PathEscape(a.Category), strings.ToLower(a.Category))
		}

		dt := fmt.Sprintf("%s ", a.DatePub.Format("Jan 2 15:04"))

		summ := trimTxt(a.ContentText)
		if len(a.Summary) > len(summ) {
			summ = a.Summary
		}
		author := ""
		if a.AuthorName != "" {
			author = fmt.Sprintf("<a href='?a=%s'>%s</a>", url.PathEscape(a.AuthorName), strings.ToLower(a.AuthorName))
		}
		element := fmt.Sprintf(article_, rd, a.Title, summ, dt,
			fmt.Sprintf("<a href='?d=%s'>%s</a>", url.PathEscape(a.Host), strings.ToLower(a.Host)),
			categ, author, px)

		items = append(items, element)
	}
	if len(items) > 300 {
		items = items[:300] //limit by 300 articles
	}
	return strings.Join(items, "")
}

func trimTxt(s string) string {
	flds := strings.Fields(s)
	if len(flds) < 150 {
		return s
	}
	bld := strings.Builder{}
	for i, str := range flds {
		bld.WriteString(str)
		bld.WriteRune(' ')
		if i >= 100 {
			bld.WriteString(" ...")
			break
		}
	}
	return bld.String()
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

func (s *Server) ArticlesByDateC(from, to time.Time, params url.Values) ([]model.Article, error) {
	art := make([]model.Article, 0)

	tx := s.DB.Where("created_at BETWEEN ? AND ?", from, to)
	if params != nil {
		if params.Get("c") != "" {
			tx = s.DB.Where("category=? AND created_at BETWEEN ? AND ?", params.Get("c"), from, to)
		}
		if params.Get("d") != "" {
			tx = s.DB.Where("host=? AND created_at BETWEEN ? AND ?", params.Get("d"), from, to)
		}
		if params.Get("a") != "" {
			tx = s.DB.Where("author_name=? AND created_at BETWEEN ? AND ?", params.Get("a"), from, to)
		}
	}
	tx.Find(&art)
	return art, tx.Error
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

// upd/ins (upsert) user
// if id not set - create new user
// in other cases save fields in db
// return cookie for update
func (s *Server) userUps(host string, usr *model.User) (*http.Cookie, error) {
	res := s.DB.Save(usr)
	if res.Error != nil {
		return nil, res.Error
	}

	b, err := json.Marshal(usr)
	if err != nil {
		return nil, err
	}
	//set cookie
	// create a Signer (HMAC in this example)
	signer, err := jwt.NewSignerHS(jwt.HS256, s.Token)
	if err != nil {
		return nil, err
	}
	// create a Builder
	builder := jwt.NewBuilder(signer)
	token, err := builder.Build(b)
	if err != nil {
		return nil, err
	}
	cookie := http.Cookie{
		Name:    "doge",
		Domain:  host,
		Value:   token.String(),
		Path:    "/",
		Expires: time.Now().Add(365 * 24 * time.Hour),
	}
	return &cookie, nil
}

// userCurr - read info about current user from cookie
func (s *Server) userCurr(cookie string) (*model.User, error) {
	usr := &model.User{}
	verifier, err := jwt.NewVerifierHS(jwt.HS256, s.Token)
	if err != nil {
		return nil, err
	}
	newToken, err := jwt.ParseAndVerifyString(cookie, verifier)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(newToken.RawClaims(), usr)
	if err != nil {
		return nil, err
	}
	return usr, nil
}
