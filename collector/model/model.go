package model

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

import (
	"github.com/realzhangm/leetcode_collector/collector/leetcode_cli"
)

type InfoNode struct {
	UserName  string `json:"user_name"`
	NumSolved int    `json:"num_solved"`
	NumTotal  int    `json:"num_total"`
	AcEasy    int    `json:"ac_easy"`
	AcMedium  int    `json:"ac_medium"`
	AcHard    int    `json:"ac_hard"`
}

type PersonInfoNode struct {
	sync.Mutex
	InfoNode `json:"info_node"`
	// key is question slug
	AcProblems       map[string]leetcode_cli.ProblemStatus               `json:"ac_problems"`
	AcProblemsDetail map[string]leetcode_cli.Question                    `json:"ac_problems_detail"`
	AcSubmissions    map[string]map[string]leetcode_cli.SubmissionDetail `json:"ac_submissions"`
}

func (p *PersonInfoNode) SetAcProblemDetail(slug string, q *leetcode_cli.Question) {
	p.Lock()
	defer p.Unlock()
	p.AcProblemsDetail[slug] = *q
}

func (p *PersonInfoNode) DeleteAcSetAcSubmission(slug string) {
	p.Lock()
	defer p.Unlock()
	delete(p.AcSubmissions[slug], slug)
}

func (p *PersonInfoNode) ProblemsDetailExist(slug string) bool {
	p.Lock()
	defer p.Unlock()
	if _, e := p.AcProblemsDetail[slug]; e {
		return true
	}
	return false
}

func (p *PersonInfoNode) SubmissionsNeedUpdate(slug string, lang string, timeStampNew string) bool {
	p.Lock()
	defer p.Unlock()
	m2, e1 := p.AcSubmissions[slug]
	if !e1 {
		return true
	}
	if _, e2 := m2[lang]; !e2 {
		return true
	}

	oldTimeStamp := strconv.FormatInt(int64(m2[lang].Timestamp), 10)
	fmt.Println(oldTimeStamp, timeStampNew)
	if strings.Compare(oldTimeStamp, timeStampNew) < 0 {
		return true
	}
	return false
}

func (p *PersonInfoNode) GetProblemsDetailExist(slug string) *leetcode_cli.Question {
	p.Lock()
	defer p.Unlock()
	if v, e := p.AcProblemsDetail[slug]; e {
		return &v
	}
	return nil
}

func (p *PersonInfoNode) GetAcSubmissions(slug string) map[string]leetcode_cli.SubmissionDetail {
	p.Lock()
	defer p.Unlock()
	m2, e1 := p.AcSubmissions[slug]
	if !e1 {
		return nil
	}
	return m2
}

func (p *PersonInfoNode) SetAcSubmissions(slug string, s *leetcode_cli.SubmissionDetail) {
	p.Lock()
	defer p.Unlock()

	if len(slug) == 0 || len(s.Lang) == 0 {
		panic("")
	}

	lang := s.Lang
	m2, e1 := p.AcSubmissions[slug]
	//fmt.Println(slug, lang, s.ID, time.Unix(int64(s.Timestamp), 0).Format("2006-01-02 03:04:05"))
	if e1 {
		// 某种语言的题解只存一份
		// 保留最新的
		if v, e2 := m2[lang]; e2 && v.Timestamp > s.Timestamp {
			return
		}
		m2[lang] = *s
	} else {
		p.AcSubmissions[slug] = make(map[string]leetcode_cli.SubmissionDetail)
		p.AcSubmissions[slug][lang] = *s
	}
}
