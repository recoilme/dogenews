package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/recoilme/dogenews/model"
	"github.com/recoilme/dogenews/web"
	"github.com/recoilme/graceful"
	"github.com/tidwall/interval"
	"golang.org/x/crypto/acme/autocert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	ver = 5 //v0.0.5
	//params
	address  = flag.String("address", ":80", "address to listen on (default: :80)")
	dbFile   = flag.String("dbfile", "db.db", "database file (main)")
	statFile = flag.String("statfile", "stat.db", "database file (stat)")
	updInt   = flag.Int("updint", 100, "interval update (seconds)")
	insInt   = flag.Int("insint", 5, "interval insert (seconds)")
)

// debug: go run main.go -address=":8080"
func main() {
	flag.Parse()
	tgtoken, err := ioutil.ReadFile("tg")
	if err != nil {
		log.Fatal(err)
	}
	tg := bytes.TrimSpace(tgtoken)
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,  // Slow SQL threshold
			LogLevel:                  logger.Error, // Log level
			IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,         // color
		},
	)
	//main db
	db, err := gorm.Open(sqlite.Open(*dbFile), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatal(err)
	}
	// migrate
	err = db.AutoMigrate(&model.Article{})
	if err != nil {
		log.Fatal(err)
	}

	// close on exit
	if sqlDB, err := db.DB(); err == nil {
		defer sqlDB.Close()
	}

	//stat db
	stat, err := gorm.Open(sqlite.Open(*statFile), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatal(err)
	}
	if ver == 5 {
		err = stat.Migrator().DropTable("events")
		if err != nil {
			log.Fatal(err)
		}
	}

	err = stat.AutoMigrate(&model.Event{})
	if err != nil {
		log.Fatal(err)
	}

	err = stat.AutoMigrate(&model.User{})
	if err != nil {
		log.Fatal(err)
	}

	if sqlDBStat, err := stat.DB(); err == nil {
		defer sqlDBStat.Close()
	}

	srv := &web.Server{DB: db, Stat: stat, Token: tg,
		Evs: &model.EventBuf{Mu: sync.Mutex{}}}

	// import
	nextCheck := time.Now()
	srv.Iv = interval.Set(func(t time.Time) {
		validate := false
		if nextCheck.Before(time.Now()) {
			validate = true
			nextCheck = time.Now().Add(1 * time.Hour)
		}
		err := srv.Import("vc_ru", validate)
		if err != nil {
			fmt.Println(err)
		}
		errH := srv.Import("habr_com", validate)
		if errH != nil {
			fmt.Println(errH)
		}

	}, time.Second*time.Duration(*updInt))

	// write stats
	srv.IvEv = interval.Set(func(t time.Time) {
		var evLen int
		var evs []*model.Event
		srv.Evs.Mu.Lock()
		evLen = len(srv.Evs.Buf)
		if evLen > 0 {
			evs = make([]*model.Event, evLen)
			copy(evs, srv.Evs.Buf)
			srv.Evs.Buf = make([]*model.Event, 0, evLen)
		}
		srv.Evs.Mu.Unlock()
		//fmt.Println(len(evs))
		if len(evs) > 0 {
			tx := stat.Create(&evs)
			if tx.Error != nil {
				fmt.Println("events", tx.Error)
			}
		}
	}, time.Second*time.Duration(*insInt))

	// signal check
	quit := make(chan os.Signal, 1)
	graceful.Unignore(quit, fallback, []os.Signal{syscall.SIGINT, syscall.SIGTERM}...)

	//web server
	if *address == ":80" {
		//run on server - redirect HTTP 2 HTTPS
		go http.ListenAndServe(*address, http.HandlerFunc(redirectHTTP))
		//run HTTP/2 server
		fmt.Println("Start:", time.Now())
		log.Fatal(http.Serve(autocert.NewListener("doge.news"), srv))
	}
	//run on localhost/debug via HTTP/1.1 (8080 and so on port)
	fmt.Println("Start(debug):", time.Now())
	log.Fatal(http.ListenAndServe(*address, srv))
}

func redirectHTTP(w http.ResponseWriter, req *http.Request) {
	target := "https://" + req.Host + req.URL.Path
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}
	//log.Printf("redirect to: %s", target)
	http.Redirect(w, req, target,
		// consider the codes 308, 302, or 301
		http.StatusTemporaryRedirect)
}

func fallback() error {
	fmt.Println("Bye", time.Now())
	return nil
}
