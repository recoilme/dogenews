package web

import (
	"fmt"
	"testing"

	"github.com/recoilme/dogenews/domains/vc_ru"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	site := vc_ru.New("vc.ru")
	links, err := site.Links()
	assert.NoError(t, err)
	fmt.Println(links)
}

func TestDel(t *testing.T) {
	link := "https://vc.ru/life/244142"
	site := vc_ru.New("vc.ru")
	_, ok := site.LinkOk(link, true)
	assert.Equal(t, false, ok)
}
