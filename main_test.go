package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wesleym/telegramwidget"
)

func Test_Auth(t *testing.T) {
	tgtoken, err := ioutil.ReadFile("tg")
	assert.NoError(t, err)
	qparams := "id=1263310&first_name=recoilme&username=recoilme&photo_url=https%3A%2F%2Ft.me%2Fi%2Fuserpic%2F320%2FjKp4n4Lk3i9yDV1dBo3WQrL3mFaQl7bgLgd0Ip_UWZM.jpg&auth_date=1621670472&hash=f0a4043c97c00d897a346e94468e92699b60643111150309e6bb6b7b9b0ae024"
	tg := strings.TrimSpace(string(tgtoken))
	params, err := url.ParseQuery(qparams)
	assert.NoError(t, err)
	u, err := telegramwidget.ConvertAndVerifyForm(params, tg)
	assert.NoError(t, err)
	assert.Equal(t, "https://t.me/i/userpic/320/jKp4n4Lk3i9yDV1dBo3WQrL3mFaQl7bgLgd0Ip_UWZM.jpg", fmt.Sprintf("%s", u.PhotoURL))
}
