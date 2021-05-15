package web

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMain(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("../test.db"), &gorm.Config{})
	assert.NoError(t, err)
	srv := &Server{DB: db}
	art, err := srv.ArticlesByDateC(time.Now().Add(time.Duration(-24)*time.Hour), time.Now())
	assert.NoError(t, err)
	for _, a := range art {
		fmt.Println(a.CreatedAt, a.DatePub)
	}
}

func TestParetto(t *testing.T) {
	for i := 0; i < 10; i++ {
		score, _ := paretto(i, 10)
		fmt.Println(i, score)
	}
}
