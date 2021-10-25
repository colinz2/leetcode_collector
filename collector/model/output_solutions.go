package model

import (
	"encoding/json"
	"fmt"
	"github.com/realzhangm/leetcode_aid/collector/leetcode_cli"
	"github.com/realzhangm/leetcode_aid/collector/util"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"
)

// 题目描述 README 中文模板
const SolutionReadme = `
# {{title_cn}}

## 题目描述

{{content_cn}}

## 题解

{{solutions}}

## 相关话题

{{tags_cn}}

## 相似题目

{{similar_questions_cn}}
`

func findExt(lang string) string {
	switch lang {
	case "python", "python3":
		return ".py"
	case "rust":
		return ".rs"
	case "golang":
		return ".go"
	}
	return "." + lang
}

func mkdir(dir string) {
	if !util.PathExists(dir) {
		os.Mkdir(dir, os.ModePerm)
	}
}

type SolutionReadMeFormatter struct {
	slug       string
	subLangMap map[string]leetcode_cli.SubmissionDetail
	question   *leetcode_cli.Question
	p          *PersonInfoNode
}

// 支持函数参数
func (s SolutionReadMeFormatter) titleCn() string {
	return fmt.Sprintf("[%s](%s%s)",
		s.question.TranslatedTitle, leetcode_cli.UrlProblems, s.slug)
}

func (s SolutionReadMeFormatter) contentCn() string {
	return s.question.TranslatedContent
}

func (s SolutionReadMeFormatter) solutions() string {
	sb := strings.Builder{}
	for lang, s := range s.subLangMap {
		sb.WriteString("### " + lang + "\n")
		sb.WriteString("```" + lang)
		sb.WriteString("\n")
		sb.WriteString(s.Code)
		sb.WriteString("\n")
		sb.WriteString("```")
		sb.WriteString("\n")
	}
	return sb.String()
}

func (s SolutionReadMeFormatter) tagsCn() string {
	res := ""
	for _, tag := range s.question.TopicTags {
		res += fmt.Sprintf("[%s](%s%s)",
			tag.TranslatedName, leetcode_cli.UrlTag, tag.Slug) + "<br>"
	}
	return res
}

type SimilarQuestion struct {
	Difficulty      string `json:"difficulty"`
	Title           string `json:"title"`
	TitleSlug       string `json:"titleSlug"`
	TranslatedTitle string `json:"translatedTitle"`
}

func (s SolutionReadMeFormatter) similarQuestionsCn() string {
	var sqs []SimilarQuestion
	json.Unmarshal([]byte(s.question.SimilarQuestions), &sqs)
	res := ""
	for _, sq := range sqs {
		if _, e := s.p.AcProblems[sq.TitleSlug]; e {
			res += fmt.Sprintf("[%s](solutions/%s/README.md) %s",
				sq.TranslatedTitle, sq.TitleSlug, sq.Difficulty) + "<br>"
		}
	}
	return res
}

func (s *SolutionReadMeFormatter) outPutSolutionReadme(slugDir string) {
	tmpl, err := template.New("all").Funcs(template.FuncMap{
		"title_cn":             (*s).titleCn,
		"content_cn":           (*s).contentCn,
		"solutions":            (*s).solutions,
		"tags_cn":              (*s).tagsCn,
		"similar_questions_cn": (*s).similarQuestionsCn,
	}).Parse(SolutionReadme)
	if err != nil {
		panic(err)
	}

	f, err := os.OpenFile(path.Join(slugDir, "README.md"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(f, s)
	if err != nil {
		panic(err)
	}
}

func (p *PersonInfoNode) OutputSolutions(outputDir string) error {
	mkdir(outputDir)

	outputOne := func(slug string) {
		question := p.GetProblemsDetailExist(slug)
		if question == nil {
			panic("slug problem " + "not exist")
		}

		subLangMap := p.GetAcSubmissions(slug)
		if subLangMap == nil {
			panic(fmt.Sprintf("slug %s subLangMap not exist", slug))
		}

		slugDir := path.Join(outputDir, slug)
		mkdir(slugDir)

		for lang, s := range subLangMap {
			srcCodeFile := path.Join(slugDir, slug+findExt(lang))
			ioutil.WriteFile(srcCodeFile, []byte(s.Code), os.ModePerm)
		}

		readMeF := SolutionReadMeFormatter{
			subLangMap: subLangMap,
			question:   question,
			slug:       slug,
			p:          p,
		}
		readMeF.outPutSolutionReadme(slugDir)
	}

	for slug, ps := range p.AcProblems {
		if slug != ps.Stat.QuestionTitleSlug {
			panic("slug not equal")
		}
		outputOne(slug)
	}
	return nil
}
