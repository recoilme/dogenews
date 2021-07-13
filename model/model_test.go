package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestModelUser(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("../db.db"), &gorm.Config{})
	assert.NoError(t, err)

	//2 new user
	usr := User{}
	res := db.Create(&usr)
	assert.NoError(t, res.Error)
	fmt.Println(usr.ID)
	usr = User{}
	res = db.Create(&usr)
	assert.NoError(t, res.Error)
	fmt.Println(usr.ID)
	db.Where(&usr).First(&usr)
	fmt.Println(usr.ID, usr.AuthDate, usr.FirstName, "tg", usr.TgId)

	//check
	id := int64(123456)
	usr = User{TgId: &id}
	tx := db.Where(&usr).First(&usr)
	//assert.NoError(t, tx.Error)
	if tx.Error == nil {
		//exists
		fmt.Println("exists", usr.ID, *usr.TgId)
	} else {
		res = db.Create(&usr)
		assert.NoError(t, res.Error)
		fmt.Println(usr.ID)
		tx := db.Where(&usr).First(&usr)
		assert.NoError(t, tx.Error)
		fmt.Println(usr.ID, *usr.TgId)
	}
}
