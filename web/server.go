package web

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/recoilme/dogenews/model"
	. "github.com/stevelacy/daz"
	"github.com/tidwall/interval"
	"gorm.io/gorm"
)

type Server struct {
	DB *gorm.DB
	Iv interval.Interval
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
		case path == "" || path == "today" || path == "yesterday" || path == "week":
			bin, err := s.Main(path)
			if checkErr(err, w) {
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(bin)
		case strings.HasPrefix(path, "import/"):
			path = strings.TrimPrefix(path, "import/")
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
			ok := checkTelegramAuthorization(path, "1705051125:AAGIcJjXyy2Bjf-Y0nQepoMV7unOBMzegAM")
			if ok {
				info, _ := json.MarshalIndent(path, "", "  ")
				cookie := http.Cookie{
					Name:    "tg",
					Domain:  "doge.news",
					Value:   string(info),
					Path:    "/",
					Expires: time.Now().Add(365 * 24 * time.Hour),
				}
				http.SetCookie(w, &cookie)
				http.Redirect(w, r, "https://doge.news", http.StatusTemporaryRedirect)
				/*w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "text/html")
				w.Write(info)*/
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		default:

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
	case "", "today":
		from = to.Add(time.Duration(-24*1) * time.Hour)
	case "yesterday":
		to = to.Add(time.Duration(-24*1) * time.Hour)
		from = to.Add(time.Duration(-24*1) * time.Hour)
	case "week":
		from = to.Add(time.Duration(-24*7) * time.Hour)
	}

	art, err := s.ArticlesByDateC(from, to)
	if err != nil {
		return nil, err
	}
	body := H(
		"body", Attr{"class": "text-gray-400 bg-gray-900 body-font"},
		H("div", UnsafeContent(Menu())),
		H("div", Attr{"class": "max-w-4xl mx-auto"}, UnsafeContent(Arts(art, path))),
	)
	html := H("html", head("doge · news"), body)
	return []byte(html()), nil
}

func Menu() string {
	return menu
}

func Arts(art []model.Article, path string) string {

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
			return int(art[i].Score*10000) > int(art[j].Score*10000)
		})
	}

	for i, a := range art {
		if i <= 1 || (i > 5 && i < 12) || (i > 21 && i < 42) {
			hero := fmt.Sprintf(artHero2, strings.ToUpper(a.Category), a.TitleMl, a.SummaryMl, a.Url,
				a.AuthorName, fmt.Sprintf("%d", a.CntComm), fmt.Sprintf("%d", a.CntLike))

			items = append(items, hero)
			continue
		}
		element := fmt.Sprintf(artBody, strings.ToUpper(a.Category), a.TitleMl, a.SummaryMl, a.Url,
			fmt.Sprintf("%s", a.ScoreTxt), fmt.Sprintf("%d", a.CntComm), fmt.Sprintf("%d", a.CntLike),
			a.AuthorAva, a.AuthorName, strings.ToUpper(a.Host))
		items = append(items, element)
	}
	if len(items) > 301 {
		items = items[:301] //limit by 300 articles
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

func checkTelegramAuthorization(data, token string) bool {
	params, _ := url.ParseQuery(data)
	strs := []string{}
	var hash = ""
	for k, v := range params {
		if k == "hash" {
			hash = v[0]
			continue
		}
		strs = append(strs, k+"="+v[0])
	}
	sort.Strings(strs)
	var imploded = ""
	for _, s := range strs {
		if imploded != "" {
			imploded += "\n"
		}
		imploded += s
	}
	sha256hash := sha256.New()
	io.WriteString(sha256hash, token)
	hmachash := hmac.New(sha256.New, sha256hash.Sum(nil))
	io.WriteString(hmachash, imploded)
	ss := hex.EncodeToString(hmachash.Sum(nil))
	return hash == ss
}
