package aid

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/machinebox/graphql"
)

const (
	Url             = "https://leetcode-cn.com"
	LoginPath       = "/accounts/login"
	SubmissionsPath = "/api/submissions"
	GraphqlPath     = "/graphql"

	UAStr = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/64.0.3282.140 Safari/537.36 Edge/17.17134"
)

func setHeader(header *http.Header) {
	header.Set("User-Agent", UAStr)
	header.Set("Origin", Url)
}

type Client struct {
	conf      ClientConf
	httpCli   *http.Client
	cookieJar *http.CookieJar
}

type ClientConf struct {
	UserName string
	PassWord string
}

// Jar https://stackoverflow.com/questions/12756782/go-http-post-and-use-cookies
type Jar struct {
	lk      sync.Mutex
	cookies map[string][]*http.Cookie
}

func NewJar() *Jar {
	jar := new(Jar)
	jar.cookies = make(map[string][]*http.Cookie)
	return jar
}

func (jar *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.lk.Lock()
	jar.cookies[u.Host] = cookies
	jar.lk.Unlock()
}

func (jar *Jar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies[u.Host]
}

func NewClient(conf *ClientConf) *Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 10 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,

			ExpectContinueTimeout: 4 * time.Second,
			ResponseHeaderTimeout: 3 * time.Second,
		},
		// Prevent endless redirects
		Timeout: 1 * time.Minute,
		Jar:     NewJar(),
	}

	return &Client{
		conf:    *conf,
		httpCli: httpClient,
	}
}

type logInParam struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (l logInParam) postFormBuffer() *bytes.Buffer {
	postFormValues := url.Values{}
	postFormValues.Add("login", l.Login)
	postFormValues.Add("password", l.Password)
	return bytes.NewBufferString(postFormValues.Encode())
}

func (c *Client) Login(ctx context.Context) error {
	loginUrl := Url + LoginPath
	logInParam := logInParam{
		Login:    c.conf.UserName,
		Password: c.conf.PassWord,
	}

	postReq, err := http.NewRequest(http.MethodPost, loginUrl, logInParam.postFormBuffer())
	if err != nil {
		return err
	}
	setHeader(&postReq.Header)
	postReq.Header.Set("referer", loginUrl)
	postReq.Header.Set("x-requested-with", "XMLHttpRequest")
	// content-type must be right: application/x-www-form-urlencoded
	postReq.Header.Set("content-type", "application/x-www-form-urlencoded")

	response, err := c.httpCli.Do(postReq)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("http status not OK, %d", response.StatusCode))
	}
	return nil
}

func (c *Client) GetSubmission(ctx context.Context) error {
	submissionUrl := Url + SubmissionsPath + "?offset=0&limit=2000"
	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, submissionUrl, nil)
	if err != nil {
		return err
	}
	setHeader(&getReq.Header)
	getReq.Header.Set("content-type", "application/json")

	getResponse, err := c.httpCli.Do(getReq)
	if err != nil {
		return err
	}
	defer getResponse.Body.Close()

	responseBody, err := io.ReadAll(getResponse.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(responseBody))

	return nil
}

func (c *Client) GraphqlTest() {
	graphqlUrl := Url + GraphqlPath
	graphqlCli := graphql.NewClient(graphqlUrl, graphql.WithHTTPClient(c.httpCli))

	req := graphql.NewRequest(`
	query questionData($titleSlug: String!) {
                question(titleSlug: $titleSlug) {
                    questionId
                    content
                    translatedTitle
                    translatedContent
                    similarQuestions
                    topicTags {
                        name
                        slug
                        translatedName
                    }
                    hints
                }
            }
	`)
	req.Var("titleSlug", "two-sum")
	var responseData map[string]interface{}
	setHeader(&req.Header)
	req.Header.Set("Cache-Control", "no-cache")
	err := graphqlCli.Run(context.TODO(), req, &responseData)
	if err != nil {
		panic(err)
	}
	fmt.Println(responseData)
}
