package model

import (
	"fmt"
	lccli "github.com/realzhangm/leetcode_collector/pkg/collector/leetcode_cli"
	"os"
	"path"
	"strings"
	"text/template"
)

// tags
// tag名.md 文件

// 题目描述 README 中文模板
const TagsMarkDown = `
# {{title_cn}}

{{tags_list}}

`

type TagFormatter struct {
	tagSlug  string
	tagLinks []TagsLink
}

func NewTagFormatter(ts string, tl []TagsLink) *TagFormatter {
	return &TagFormatter{
		tagSlug:  ts,
		tagLinks: tl,
	}
}

// 支持函数参数
func (t *TagFormatter) titleCn() string {
	return fmt.Sprintf("[%s](%s%s)",
		t.tagLinks[0].topicTag.TranslatedName, lccli.UrlTag, t.tagLinks[0].topicTag.Slug)
}

func (t *TagFormatter) tagsList() string {
	sb := strings.Builder{}

	for i, tagLink := range t.tagLinks {
		sb.WriteString(fmt.Sprintf("%d. ", i+1))
		sb.WriteString(fmt.Sprintf("[%s](../solutions/%s/README.md)",
			tagLink.question.TranslatedTitle, tagLink.problemStatus.Stat.QuestionTitleSlug))
		sb.WriteString("\n")
	}
	return sb.String()
}

func (t *TagFormatter) outPutTagMarkDown(tagsDir string) {
	tmpl, err := template.New("tag").Funcs(template.FuncMap{
		"title_cn":  (*t).titleCn,
		"tags_list": (*t).tagsList,
	}).Parse(TagsMarkDown)
	if err != nil {
		panic(err)
	}

	fileName := t.tagSlug + ".md"
	f, err := os.OpenFile(path.Join(tagsDir, fileName), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(f, t)
	if err != nil {
		panic(err)
	}
}

func (p *PersonInfoNode) OutputTags(outputDir string) error {
	mkdir(outputDir)
	for slug, tagLinks := range p.TagsMap {
		NewTagFormatter(slug, tagLinks).outPutTagMarkDown(outputDir)
	}
	return nil
}
