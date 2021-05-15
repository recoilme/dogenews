package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/recoilme/dogenews/model"
	"github.com/recoilme/dogenews/web"
	"github.com/tidwall/interval"
	"golang.org/x/crypto/acme/autocert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	//TODO params
	address := ""
	dbFile := "db.db"
	updInt := 100
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,  // Slow SQL threshold
			LogLevel:                  logger.Error, // Log level
			IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
			Colorful:                  true,         // color
		},
	)
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = db.AutoMigrate(&model.Article{})
	if err != nil {
		log.Fatal(err)
	}

	srv := &web.Server{DB: db}
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

	}, time.Second*time.Duration(updInt))

	if address == "" {
		log.Fatal(http.Serve(autocert.NewListener("doge.news"), srv))
	} else {
		log.Fatal(http.ListenAndServe(address, srv))
	}
}
