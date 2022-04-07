package model

import (
	"fmt"
	"github.com/realzhangm/leetcode_collector/pkg/doa"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

import (
	"github.com/realzhangm/leetcode_collector/pkg/bufferpool"
	lccli "github.com/realzhangm/leetcode_collector/pkg/collector/leetcode_cli"
)

const markdownTemplate = `
<p align="center"><img width="300" src="https://raw.githubusercontent.com/KivenCkl/LeetCode_Helper/master/imgs/leetcode-logo.png"></p>
<p align="center">
    <img src="https://img.shields.io/badge/用户-{{.InfoNode.UserName}}-blue.svg?" alt="">
    <img src="https://img.shields.io/badge/已解决-{{.InfoNode.NumSolved}}/{{.InfoNode.NumTotal}}-blue.svg?" alt="">
    <img src="https://img.shields.io/badge/简单-{{.InfoNode.AcEasy}}-green.svg?" alt="">
    <img src="https://img.shields.io/badge/中等-{{.InfoNode.AcMedium}}-orange.svg?" alt="">
    <img src="https://img.shields.io/badge/困难-{{.InfoNode.AcHard}}-red.svg?" alt="">
</p>
<h1 align="center">LeetCode 的解答</h1>

<p align="center">
    <br>
    <b>最近一次更新: {{Time}} </b>
    <br>
</p>
<!--请保留下面这行信息，让更多用户了解到这个小爬虫，衷心感谢您的支持-->
<p align="center">The source code is fetched using the tool <a href="https://github.com/realzhangm/leetcode_collector">leetcode_collector</a>.</p>

<p align="center"> 查看 <a href="./TAGS.md">标签视角</a>.</p>


| # | 题名 | 解答 | 通过率 | 难度 | 标签 |
|:--:|:-----|:---------:|:----:|:----:|:----:|
{{Summary}}
`
const TableLine1 = `|{frontend_id}|{title}{paid_only}{is_favor}|{solutions}|{ac_rate}|{difficulty}|{tags}|`

type TableLineFormat struct {
	slug string
	ps   *lccli.ProblemStatus
	q    *lccli.Question
}

func (t TableLineFormat) frontendId() string {
	// 真正的序号
	return t.ps.Stat.FrontendQuestionID
}

func (t TableLineFormat) acRate() string {
	r := float64(t.ps.Stat.TotalAcs) * 100 / float64(t.ps.Stat.TotalSubmitted)
	return strconv.FormatFloat(r, 'f', 1, 64) + "%"
}

func (t TableLineFormat) solutions() string {
	return fmt.Sprintf("[🔗](solutions/%s/README.md)", t.slug)
}

func (t TableLineFormat) title() string {
	return fmt.Sprintf("[%s](%s%s)",
		t.q.TranslatedTitle, lccli.UrlProblems, t.ps.Stat.QuestionTitleSlug)
}

func (t TableLineFormat) paidOnly() string {
	if t.ps.PaidOnly {
		return " 🔒"
	}
	return ""
}

func (t TableLineFormat) isFavor() string {
	if t.ps.IsFavor {
		return " ♥"
	}
	return ""
}

func (t TableLineFormat) difficulty() string {
	switch t.ps.Difficulty.Level {
	case 1:
		return "简单"
	case 2:
		return "中等"
	case 3:
		return "困难"
	}
	return "未知"
}
func (t TableLineFormat) tags() string {
	res := ""
	for _, tag := range t.q.TopicTags {
		res += fmt.Sprintf("[%s](%s/%s.md)",
			tag.TranslatedName, "tags", tag.Slug) + "<br>"
	}
	return res
}

func (t *TableLineFormat) templateExe() string {
	tmpl, err := template.New("table_line").Delims("{", "}").Funcs(template.FuncMap{
		"frontend_id": (*t).frontendId,
		"title":       (*t).title,
		"paid_only":   (*t).paidOnly,
		"is_favor":    (*t).isFavor,
		"solutions":   (*t).solutions,
		"ac_rate":     (*t).acRate,
		"difficulty":  (*t).difficulty,
		"tags":        (*t).tags,
	}).Parse(TableLine1)
	if err != nil {
		panic(err)
	}

	buffer := bufferpool.GetBuffer()
	defer bufferpool.PutBuffer(buffer)
	err = tmpl.Execute(buffer, t)
	if err != nil {
		panic(err)
	}
	buffer.WriteString("\n")
	return buffer.String()
}

type SubmissionDetailSlice []lccli.SubmissionDetail

func (s SubmissionDetailSlice) Len() int      { return len(s) }
func (s SubmissionDetailSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s SubmissionDetailSlice) Less(i, j int) bool {
	return strings.Compare(s[i].Lang, s[i].Lang) < 0
}

type ProblemStatusSlice []lccli.ProblemStatus

func (s ProblemStatusSlice) Len() int      { return len(s) }
func (s ProblemStatusSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ProblemStatusSlice) Less(i, j int) bool {
	// 按照问题序号排序，由高到低
	return s[i].Stat.QuestionID < s[j].Stat.QuestionID
}

func (p *PersonInfoNode) summaryTable() string {
	bd := strings.Builder{}
	pSlice := ProblemStatusSlice{}
	for slug, ps := range p.AcProblems {
		if slug != ps.Stat.QuestionTitleSlug {
			panic("slug not equal")
		}
		pSlice = append(pSlice, ps)
	}
	sort.Sort(pSlice)

	for _, ps := range pSlice {
		slug := ps.Stat.QuestionTitleSlug
		pd, e := p.AcProblemsDetail[slug]
		if !e {
			panic("not exist ?")
		}

		doa.Assert(len(pd.TranslatedTitle) > 0)

		tlf := TableLineFormat{
			slug: slug,
			ps:   &ps,
			q:    &pd,
		}
		bd.WriteString(tlf.templateExe())
	}
	return bd.String()
}

func updateTime() string {
	return time.Now().Format("2006年1月2日 15:04:05")
}

func (p *PersonInfoNode) json2MD1(outputDir string) error {
	tmpl, err := template.New("all").Funcs(template.FuncMap{
		"Time":    updateTime,
		"Summary": (*p).summaryTable,
	}).Parse(markdownTemplate)
	if err != nil {
		panic(err)
	}

	f, err := os.OpenFile(outputDir, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = tmpl.Execute(f, p)
	if err != nil {
		panic(err)
	}
	return nil
}

func (p *PersonInfoNode) Json2Md(outputDir string) error {
	err := p.json2MD1(outputDir)
	if err != nil {
		return err
	}
	return nil
}
