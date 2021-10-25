package leetcode_cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/realzhangm/leetcode_collector/collector/bufferpool"
	"github.com/realzhangm/leetcode_collector/collector/util"
)

const (
	Url                 = "https://leetcode-cn.com"
	UrlTag              = Url + "/tag/"
	UrlProblems         = Url + "/problems/"
	LoginPath           = "/accounts/login"
	SubmissionsPath     = "/api/submissions"
	GraphqlPath         = "/graphql"
	ProblemsAll         = "/api/problems/all"
	SubmissionLatestUri = "/submissions/latest/?qid=%d&lang=%s"

	UAStr = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/64.0.3282.140 Safari/537.36 Edge/17.17134"
)

var (
	ErrorClient               = errors.New("leetcode client error")
	ErrorClientLogin          = errors.Wrap(ErrorClient, "login")
	ErrorClientGraphQl        = errors.Wrap(ErrorClient, "graphql")
	ErrorClientGetAllProblems = errors.Wrap(ErrorClient, "get all problems")
	ErrorSubmissionDetail     = errors.Wrap(ErrorClient, "query submission detail")
)

func setHeader(header *http.Header) {
	header.Set("User-Agent", UAStr)
	header.Set("Origin", Url)
	header.Set("Cache-Control", "no-cache")
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

func (jar *Jar) Clone() *Jar {
	jar.lk.Lock()
	defer jar.lk.Unlock()

	j2 := NewJar()
	// map copy is pointer
	for k, v := range jar.cookies {
		j2.cookies[k] = v
	}
	return j2
}

func (jar *Jar) Print() {
	jar.lk.Lock()
	defer jar.lk.Unlock()
	for k, v := range jar.cookies {
		fmt.Println(k, v)
	}
}

func (jar *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.lk.Lock()
	jar.cookies[u.Host] = cookies
	jar.lk.Unlock()
}

func (jar *Jar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies[u.Host]
}

type Client struct {
	conf        ClientConf
	httpCli     *http.Client
	cookieJar   *Jar
	loginFlag   bool
	httpCliPool *sync.Pool
}

type ClientConf struct {
	UserName  string
	PassWord  string
	OutputDir string
}

func newGraphQlHttpClient() *http.Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 10 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,

			ExpectContinueTimeout: 4 * time.Second,
			ResponseHeaderTimeout: 3 * time.Second,
			MaxIdleConns:          2,
			MaxConnsPerHost:       2,
			MaxIdleConnsPerHost:   2,
			IdleConnTimeout:       30,
		},
		// Prevent endless redirects
		Timeout: 1 * time.Minute,
	}
	return httpClient
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
			MaxIdleConns:          10,
			MaxConnsPerHost:       5,
			MaxIdleConnsPerHost:   5,
			IdleConnTimeout:       10,
		},
		// Prevent endless redirects
		Timeout: 1 * time.Minute,
	}

	client := &Client{
		conf:      *conf,
		httpCli:   httpClient,
		cookieJar: NewJar(),
		httpCliPool: &sync.Pool{
			New: func() interface{} {
				return newGraphQlHttpClient()
			},
		},
	}

	if len(client.conf.OutputDir) == 0 {
		client.conf.OutputDir = "./output"
		util.Mkdir(client.conf.OutputDir)
	}
	return client
}

func (c *Client) getHttpClintFromPool() *http.Client {
	return c.httpCliPool.Get().(*http.Client)
}

func (c *Client) putHttpClintToPool(httpCli *http.Client) {
	httpCli.CloseIdleConnections()
	c.httpCliPool.Put(httpCli)
}

func (c *Client) isLogin() bool {
	return c.loginFlag
}

func (c *Client) setLoginFlag() {
	c.loginFlag = true
}

type LogInParam struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (l LogInParam) postFormBuffer() *bytes.Buffer {
	postFormValues := url.Values{}
	postFormValues.Add("login", l.Login)
	postFormValues.Add("password", l.Password)
	return bytes.NewBufferString(postFormValues.Encode())
}

func (c *Client) Login(ctx context.Context) error {
	loginUrl := Url + LoginPath
	logInParam := LogInParam{
		Login:    c.conf.UserName,
		Password: c.conf.PassWord,
	}

	postReq, err := http.NewRequest(http.MethodPost, loginUrl, logInParam.postFormBuffer())
	if err != nil {
		return errors.Wrap(ErrorClientLogin, err.Error())
	}
	setHeader(&postReq.Header)
	postReq.Header.Set("referer", loginUrl)
	postReq.Header.Set("x-requested-with", "XMLHttpRequest")
	// content-type must be right: application/x-www-form-urlencoded
	postReq.Header.Set("content-type", "application/x-www-form-urlencoded")
	c.httpCli.Jar = c.cookieJar

	response, err := c.httpCli.Do(postReq)
	if err != nil {
		return errors.Wrap(ErrorClientLogin, err.Error())
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("http status not OK, %d", response.StatusCode)
		return errors.Wrap(ErrorClientLogin, msg)
	}
	c.setLoginFlag()
	return nil
}

func (c *Client) graphqlURl() string {
	return Url + GraphqlPath
}

func (c *Client) problemAllURL() string {
	return Url + ProblemsAll
}

// GetAllProblems Get All the problems
func (c *Client) GetAllProblems() (error, *AllProblemsResponse) {
	if !c.isLogin() {
		return errors.Wrap(ErrorClientGetAllProblems, "not login"), nil
	}

	req, err := http.NewRequest(http.MethodGet, c.problemAllURL(), nil)
	if err != nil {
		return errors.Wrap(ErrorClientGetAllProblems, err.Error()), nil
	}

	setHeader(&req.Header)
	rsp, err := c.httpCli.Do(req)
	if err != nil {
		return errors.Wrap(ErrorClientGetAllProblems, err.Error()), nil
	}

	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("http status code:%d", rsp.StatusCode)
		return errors.Wrap(ErrorClientGetAllProblems, msg), nil
	}
	allProblemsResponse := &AllProblemsResponse{}
	dec := json.NewDecoder(rsp.Body)
	err = dec.Decode(allProblemsResponse)
	if err != nil {
		return errors.Wrap(ErrorClientGetAllProblems, err.Error()), nil
	}

	// 这里可以支持保持文件
	f, err := os.OpenFile("all_problems.json", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	buff := bufferpool.GetBuffer()
	defer bufferpool.PutBuffer(buff)
	en := json.NewEncoder(buff)
	en.SetIndent("", " ")
	err = en.Encode(allProblemsResponse)
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(f, buff)
	if err != nil {
		panic(err)
	}

	return nil, allProblemsResponse
}

// QueryQuestionDetail get every question detail
func (c *Client) QueryQuestionDetail(questionSlug string) (error, *QuestionDetailResponse) {
	if !c.isLogin() {
		return errors.Wrap(ErrorClientGetAllProblems, "not login"), nil
	}
	// don't need to change http client ?
	// must change to support multi route
	httpCli := c.getHttpClintFromPool()
	defer c.putHttpClintToPool(httpCli)
	httpCli.Jar = c.cookieJar.Clone()

	graphqlCli := graphql.NewClient(c.graphqlURl(), graphql.WithHTTPClient(httpCli))
	req := graphql.NewRequest(QueryQuestion)
	req.Var("titleSlug", questionSlug)

	var responseData map[string]interface{}
	err := graphqlCli.Run(context.TODO(), req, &responseData)
	if err != nil {
		msg := fmt.Sprintf("QueryQuestionDetail:%s", err.Error())
		err = errors.Wrap(ErrorClientGraphQl, msg)
	}

	r := &QuestionDetailResponse{}
	buff := bufferpool.GetBuffer()
	defer bufferpool.PutBuffer(buff)
	enc := json.NewEncoder(buff)
	err = enc.Encode(responseData)
	if err != nil {
		return errors.Wrap(ErrorClientGraphQl, err.Error()), nil
	}
	dec := json.NewDecoder(buff)
	err = dec.Decode(r)
	if err != nil {
		return errors.Wrap(ErrorClientGraphQl, err.Error()), nil
	}
	return nil, r
}

// QuerySubmissionsByQuestion get all the submission for each question
func (c *Client) QuerySubmissionsByQuestion(questionSlug string) (error, *SubmissionsByQuestionResponse) {
	if len(questionSlug) == 0 {
		return errors.Wrap(ErrorClientGraphQl, "questionSlug is zero length"), nil
	}

	if !c.isLogin() {
		return errors.Wrap(ErrorClientGraphQl, "not login"), nil
	}

	httpCli := c.getHttpClintFromPool()
	defer c.putHttpClintToPool(httpCli)
	httpCli.Jar = c.cookieJar.Clone()

	graphqlCli := graphql.NewClient(c.graphqlURl(), graphql.WithHTTPClient(httpCli))
	graphqlCli.Log = func(s string) {
		//log.Println(s)
	}

	req := graphql.NewRequest(QuerySubmissionByQuestionSlug)
	req.Var("questionSlug", questionSlug)
	req.Var("offset", 0)
	req.Var("limit", 100)
	setHeader(&req.Header)

	var responseData map[string]interface{}
	err := graphqlCli.Run(context.TODO(), req, &responseData)
	if err != nil {
		return errors.Wrap(ErrorClientGraphQl, err.Error()), nil
	}
	saber := &SubmissionsByQuestionResponse{}
	buff := bufferpool.GetBuffer()
	defer bufferpool.PutBuffer(buff)
	enc := json.NewEncoder(buff)
	err = enc.Encode(responseData)
	if err != nil {
		return errors.Wrap(ErrorClientGraphQl, err.Error()), nil
	}
	dec := json.NewDecoder(buff)
	err = dec.Decode(saber)
	if err != nil {
		return errors.Wrap(ErrorClientGraphQl, err.Error()), nil
	}
	return nil, saber
}

// QuerySubmissionDetail get very submssion detail
func (c *Client) QuerySubmissionDetail(id int64) (error, *SubmissionDetailResponse) {
	if !c.isLogin() {
		return errors.Wrap(ErrorSubmissionDetail, "not login"), nil
	}

	// don't need to change http client ?
	httpCli := c.getHttpClintFromPool()
	defer c.putHttpClintToPool(httpCli)
	httpCli.Jar = c.cookieJar.Clone()

	graphqlCli := graphql.NewClient(c.graphqlURl(), graphql.WithHTTPClient(httpCli))
	graphqlCli.Log = func(s string) {
		log.Println(s)
	}

	req := graphql.NewRequest(QuerySubmissionDetail)
	req.Var("id", id)
	setHeader(&req.Header)

	//(httpCli.Jar.(*Jar)).Print()

	var responseData map[string]interface{}
	err := graphqlCli.Run(context.TODO(), req, &responseData)
	if err != nil {
		return errors.Wrap(ErrorSubmissionDetail, err.Error()), nil
	}

	r := &SubmissionDetailResponse{}
	buff := bufferpool.GetBuffer()
	defer bufferpool.PutBuffer(buff)
	enc := json.NewEncoder(buff)
	err = enc.Encode(responseData)
	if err != nil {
		return errors.Wrap(ErrorSubmissionDetail, err.Error()), nil
	}
	dec := json.NewDecoder(buff)
	err = dec.Decode(r)
	if err != nil {
		return errors.Wrap(ErrorSubmissionDetail, err.Error()), nil
	}

	if r.SubmissionDetail == nil {
		return errors.Wrap(ErrorSubmissionDetail, "body SubmissionDetail is null"), nil
	}
	return nil, r
}
