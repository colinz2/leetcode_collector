package aid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"log"
	"net/http"
)

func Scraping(conf *ClientConf) {
	c := colly.NewCollector()

	c.OnError(func(response *colly.Response, err error) {
		fmt.Println("on error:", err, response.StatusCode)
		fmt.Println(string(response.Body))
	})

	logInParam := logInParam{
		Login:    conf.UserName,
		Password: conf.PassWord,
	}
	reqBody, err := json.Marshal(&logInParam)
	if err != nil {
		panic(err)
	}

	// authenticate
	loginUrl := Url + LoginPath
	header := http.Header{}
	header.Set("Referer", loginUrl)
	err = c.Request(http.MethodPost, loginUrl, bytes.NewReader(reqBody), nil, header)
	if err != nil {
		log.Fatal("error:", err)
	}

	// attach callbacks after login
	c.OnResponse(func(r *colly.Response) {
		log.Println("response received", r.StatusCode)
	})

	// start scraping
	c.Visit(Url + SubmissionsPath)
}
