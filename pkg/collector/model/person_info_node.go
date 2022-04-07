package model

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)
import (
	lccli "github.com/realzhangm/leetcode_collector/pkg/collector/leetcode_cli"
)

type InfoNode struct {
	UserName  string `json:"user_name"`
	NumSolved int    `json:"num_solved"`
	NumTotal  int    `json:"num_total"`
	AcEasy    int    `json:"ac_easy"`
	AcMedium  int    `json:"ac_medium"`
	AcHard    int    `json:"ac_hard"`
}

type TagsLink struct {
	topicTag      lccli.TopicTag
	question      *lccli.Question
	problemStatus *lccli.ProblemStatus
}

type PersonInfoNode struct {
	sync.Mutex
	InfoNode `json:"info_node"`
	// key is question slug
	AcProblems       map[string]lccli.ProblemStatus               `json:"ac_problems"`
	AcProblemsDetail map[string]lccli.Question                    `json:"ac_problems_detail"`
	AcSubmissions    map[string]map[string]lccli.SubmissionDetail `json:"ac_submissions"`
	// tags Map
	TagsMap map[string][]TagsLink
}

func NewPersonInfoNode() *PersonInfoNode {
	return &PersonInfoNode{
		Mutex:            sync.Mutex{},
		InfoNode:         InfoNode{},
		AcProblems:       make(map[string]lccli.ProblemStatus),
		AcProblemsDetail: make(map[string]lccli.Question),
		AcSubmissions:    make(map[string]map[string]lccli.SubmissionDetail),
		TagsMap:          make(map[string][]TagsLink),
	}
}

// InsertTagsMap : 多个 tag 可以对应一个题目
func (p *PersonInfoNode) InsertTagsMap(q *lccli.Question, ps *lccli.ProblemStatus) {
	for i := range q.TopicTags {
		tagsSlug := q.TopicTags[i].Slug
		p.TagsMap[tagsSlug] = append(p.TagsMap[tagsSlug], TagsLink{
			topicTag:      q.TopicTags[i],
			question:      q,
			problemStatus: ps,
		})
	}
}

func (p *PersonInfoNode) SetAcProblemDetail(slug string, q *lccli.Question) {
	p.Lock()
	p.AcProblemsDetail[slug] = *q
	p.Unlock()
}

func (p *PersonInfoNode) DeleteAcSetAcSubmission(slug string) {
	p.Lock()
	delete(p.AcSubmissions[slug], slug)
	p.Unlock()
}

func (p *PersonInfoNode) ProblemsDetailExist(slug string) bool {
	var exist bool
	p.Lock()
	detail, exist := p.AcProblemsDetail[slug]
	if len(detail.Content) == 0 || len(detail.TranslatedTitle) == 0 {
		exist = false
	}
	p.Unlock()
	return exist
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
	fmt.Println("oldTimeStamp = ", oldTimeStamp, ", timeStampNew = ", timeStampNew)
	if strings.Compare(oldTimeStamp, timeStampNew) < 0 {
		return true
	}
	return false
}

func (p *PersonInfoNode) GetProblemsDetailExist(slug string) *lccli.Question {
	p.Lock()
	defer p.Unlock()
	if v, e := p.AcProblemsDetail[slug]; e {
		return &v
	}
	return nil
}

func (p *PersonInfoNode) GetAcSubmissions(slug string) map[string]lccli.SubmissionDetail {
	p.Lock()
	defer p.Unlock()
	m2, e1 := p.AcSubmissions[slug]
	if !e1 {
		return nil
	}
	return m2
}

func (p *PersonInfoNode) SetAcSubmissions(slug string, s *lccli.SubmissionDetail) {
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
		p.AcSubmissions[slug] = make(map[string]lccli.SubmissionDetail)
		p.AcSubmissions[slug][lang] = *s
	}
}
