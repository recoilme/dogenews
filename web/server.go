package web

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/recoilme/links/model"
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
		minimal := 300
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
	//sort by pubdate
	sort.Slice(art, func(i, j int) bool {
		if int64(art[i].Score*100) == int64(art[j].Score*100) {
			return art[i].DatePub.Unix() > art[j].DatePub.Unix()
		}
		return art[i].Score > art[j].Score
	})

	for i, a := range art {
		if i <= 1 && path != "" {
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

	items = append(items, artFoot)
	return strings.Join(items, "")
}

func paretto(i, len int) (float64, string) {
	score := float64(i+1) / float64(len)
	//fmt.Println((i), score)
	switch {
	case score <= 0.2:
		return 0.8, "A"
	case score <= 0.5:
		return 0.15, "B"
	case score <= 0.95:
		return 0.04, "C"
	default:
		return 0.01, "D"
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