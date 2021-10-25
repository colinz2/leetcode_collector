package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/realzhangm/leetcode_collector/collector/leetcode_cli"
	"github.com/realzhangm/leetcode_collector/collector/model"
	"github.com/realzhangm/leetcode_collector/collector/util"
	"golang.org/x/sync/errgroup"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

var (
	ErrCollector = errors.New("collect error")
	ErrExtractor = errors.Wrap(ErrCollector, "extracting error")
)

type Collector struct {
	ltClit *leetcode_cli.Client
	conf   Config

	personInfo model.PersonInfoNode
}

func NewCollector(c *Config) *Collector {
	collector := &Collector{
		ltClit: leetcode_cli.NewClient(&c.ltClientConf),
		conf:   *c,
		personInfo: model.PersonInfoNode{
			Mutex:            sync.Mutex{},
			InfoNode:         model.InfoNode{},
			AcProblems:       make(map[string]leetcode_cli.ProblemStatus),
			AcProblemsDetail: make(map[string]leetcode_cli.Question),
			AcSubmissions:    make(map[string]map[string]leetcode_cli.SubmissionDetail),
		},
	}
	return collector
}

func (c *Collector) fetchAcProblemsDetail() error {
	// errgroup is better
	const reqRoutineNum = 5
	slugChan := make(chan string, 1)
	g, ctx := errgroup.WithContext(context.TODO())

	g.Go(func() error {
		defer close(slugChan)
		for slug := range c.personInfo.AcProblems {
			// 这里先判断一下是否已将存在
			if c.personInfo.ProblemsDetailExist(slug) {
				continue
			}
			select {
			case slugChan <- slug:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})

	tmpMap := make(map[string]*leetcode_cli.Question)
	for i := 0; i < reqRoutineNum; i++ {
		g.Go(func() error {
			for slug := range slugChan {
				ee, q := c.ltClit.QueryQuestionDetail(slug)
				if ee != nil {
					return ee
				}
				tmpMap[slug] = &q.Question
			}
			return nil
		})
	}

	err := g.Wait()
	if err != nil {
		return err
	}

	for k, v := range tmpMap {
		c.personInfo.SetAcProblemDetail(k, v)
	}
	return nil
}

func (c *Collector) fetchAllProblems() error {
	err := c.ltClit.Login(context.TODO())
	if err != nil {
		return errors.Wrap(ErrCollector, err.Error())
	}

	err, allProblems := c.ltClit.GetAllProblems()
	if err != nil {
		return errors.Wrap(ErrCollector, err.Error())
	}
	if allProblems.UserName != c.conf.ltClientConf.UserName {
		msg := fmt.Sprintf("user name[%s]is not right", allProblems.UserName)
		return errors.Wrap(ErrCollector, msg)
	}

	c.personInfo.UserName = allProblems.UserName
	c.personInfo.NumTotal = allProblems.NumTotal
	c.personInfo.NumSolved = allProblems.NumSolved
	c.personInfo.AcEasy = allProblems.AcEasy
	c.personInfo.AcHard = allProblems.AcHard
	c.personInfo.AcMedium = allProblems.AcMedium

	for _, s := range allProblems.StatStatusPairs {
		if s.IsAc() {
			c.personInfo.AcProblems[s.Stat.QuestionTitleSlug] = s
		}
	}

	// Get question detail
	err = c.fetchAcProblemsDetail()
	if err != nil {
		return err
	}
	return nil
}

// 重试 3 次
func tryNTimes(n int, f func(i int) error) error {
	var e error
	for i := 0; i < n; i++ {
		e = f(i)
		if e == nil {
			break
		}
	}
	return e
}

// 从sbl中选择
func (c *Collector) submissionForOneLang(sbl []leetcode_cli.Submission) map[string]leetcode_cli.Submission {
	langSubmissionMap := make(map[string]leetcode_cli.Submission)
	for _, sb := range sbl {
		v, e := langSubmissionMap[sb.Lang]
		if !e || strings.Compare(v.Timestamp, sb.Timestamp) < 0 {
			if sb.StatusDisplay == "Accepted" {
				langSubmissionMap[sb.Lang] = sb
			}
		}
	}
	return langSubmissionMap
}

func (c *Collector) fetchAllSubmissions() error {
	for slug := range c.personInfo.AcProblems {
		e, sbs := c.ltClit.QuerySubmissionsByQuestion(slug)
		if e != nil {
			fmt.Println(e)
			return e
		}

		langSubmissionMap := c.submissionForOneLang(sbs.SubmissionList.Submissions)
		for _, sb := range langSubmissionMap {
			id, e2 := strconv.ParseInt(sb.ID, 10, 64)
			if e2 != nil {
				fmt.Println(e2)
				return e2
			}
			if !c.personInfo.SubmissionsNeedUpdate(slug, sb.Lang, sb.Timestamp) {
				continue
			}

			if err := tryNTimes(3, func(i int) error {
				e3, sbDetail := c.ltClit.QuerySubmissionDetail(id)
				if e3 != nil {
					fmt.Println(i, "error:", e3)
					return e3
				}
				titleSlug := sbDetail.SubmissionDetail.Question.TitleSlug
				c.personInfo.SetAcSubmissions(titleSlug, sbDetail.SubmissionDetail)
				return nil
			}); err != nil {
				// delete this titleSlug
				c.personInfo.DeleteAcSetAcSubmission(slug)
				return err
			}
		}
	}
	return nil
}

// 存在服务拒绝
func (c *Collector) fetchAllSubmissions2() error {
	g, ctx := errgroup.WithContext(context.TODO())

	slugChan := make(chan string)
	submissionsChan := make(chan *leetcode_cli.SubmissionsByQuestionResponse)

	g.Go(func() error {
		defer close(slugChan)
		for slug := range c.personInfo.AcProblems {
			select {
			case slugChan <- slug:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})

	g.Go(func() error {
		defer close(submissionsChan)
		for slug := range slugChan {
			ee, sbs := c.ltClit.QuerySubmissionsByQuestion(slug)
			if ee != nil {
				fmt.Println(ee)
				return ee
			}
			select {
			case submissionsChan <- sbs:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})

	g.Go(func() error {
		for sbs := range submissionsChan {
			langSubmissionMap := make(map[string]leetcode_cli.Submission)
			for _, sb := range sbs.SubmissionList.Submissions {
				v, e := langSubmissionMap[sb.Lang]
				if !e || strings.Compare(v.Timestamp, sb.Timestamp) < 0 {
					if sb.StatusDisplay == "Accepted" {
						langSubmissionMap[sb.Lang] = sb
					}
				}
			}

			for _, sb := range langSubmissionMap {
				id, ee := strconv.ParseInt(sb.ID, 10, 64)
				if ee != nil {
					fmt.Println(ee)
					return ee
				}
				ee, sbDetail := c.ltClit.QuerySubmissionDetail(id)
				if ee != nil {
					fmt.Println(ee)
					return ee
				}
				titleSlug := sbDetail.SubmissionDetail.Question.TitleSlug
				c.personInfo.SetAcSubmissions(titleSlug, sbDetail.SubmissionDetail)
				// have to sleep ?
			}
		}
		return nil
	})

	err := g.Wait()
	if err != nil {
		return err
	}
	return nil
}

func (c *Collector) LoadInfo() error {
	err := c.loadInfo()
	if err != nil {
		return err
	}
	return nil
}

func (c *Collector) FetchFromLeetCode() error {
	err := c.fetchAllProblems()
	if err != nil {
		return err
	}

	err = c.fetchAllSubmissions()
	if err != nil {
		return c.dumpInfo()
	}
	return c.dumpInfo()
}

func (c *Collector) allInfoFilePath() string {
	return path.Join(c.conf.OutputDir, "all_info.json")
}

func (c *Collector) loadInfo() error {
	if !util.PathExists(c.allInfoFilePath()) {
		return nil
	}

	f, err := os.OpenFile(c.allInfoFilePath(), os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	en := json.NewDecoder(f)
	err = en.Decode(&c.personInfo)
	if err != nil {
		return err
	}
	fmt.Println(len(c.personInfo.AcProblems),
		len(c.personInfo.AcProblemsDetail),
		len(c.personInfo.AcSubmissions))
	return nil
}

func (c *Collector) dumpInfo() error {
	f, err := os.OpenFile(c.allInfoFilePath(), os.O_TRUNC|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	en := json.NewEncoder(f)
	en.SetIndent("", " ")
	err = en.Encode(&c.personInfo)
	if err != nil {
		return err
	}
	return nil
}

func (c *Collector) ExtractOneMarkDown() error {
	for slug, v := range c.personInfo.AcProblems {
		fmt.Printf("slug=%s %v %s \n", slug, v.IsFavor, c.personInfo.AcProblemsDetail[slug].QuestionID)
	}

	for slug, v := range c.personInfo.AcProblemsDetail {
		fmt.Printf("slugPD=%s %v \n", slug, v.QuestionID)
	}

	fmt.Println("len(c.personInfo.AcProblemsDetail)=", len(c.personInfo.AcProblemsDetail))

	return nil
}

func (c *Collector) Json2MD() error {
	mdPath := path.Join(c.conf.OutputDir, "README.md")
	return c.personInfo.Json2Md(mdPath)
}

func (c *Collector) OutputSolutions() error {
	return c.personInfo.OutputSolutions(c.conf.SolutionsDir)
}
