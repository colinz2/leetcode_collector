package model

import (
	"fmt"
	"github.com/realzhangm/leetcode_collector/pkg/bufferpool"
	lccli "github.com/realzhangm/leetcode_collector/pkg/collector/leetcode_cli"
	"github.com/realzhangm/leetcode_collector/pkg/doa"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"
)

// tags
// tag名.md 文件
//TODO 优化

// 题目描述 README 中文模板
const TagsMarkDown = `
## {{title_cn}}

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
	tagName := t.tagLinks[0].topicTag.TranslatedName
	if len(tagName) == 0 {
		tagName = t.tagLinks[0].topicTag.Name
		tagName = strings.ToTitle(tagName)
	}

	return fmt.Sprintf("[%s](%s%s)",
		tagName, lccli.UrlTag, t.tagLinks[0].topicTag.Slug)
}

func (t *TagFormatter) tagsList() string {
	sb := strings.Builder{}

	for i, tagLink := range t.tagLinks {
		title := tagLink.question.TranslatedTitle
		if len(title) == 0 {
			title = tagLink.problemStatus.Stat.QuestionTitle
			title = strings.ToTitle(title)
		}

		sb.WriteString(fmt.Sprintf("%d. ", i+1))
		sb.WriteString(fmt.Sprintf("[%s](../solutions/%s/README.md)",
			title, tagLink.problemStatus.Stat.QuestionTitleSlug))
		sb.WriteString("\n")
	}
	return sb.String()
}

func (t *TagFormatter) String() string {
	tmpl, err := template.New("tag").Funcs(template.FuncMap{
		"title_cn":  (*t).titleCn,
		"tags_list": (*t).tagsList,
	}).Parse(TagsMarkDown)
	if err != nil {
		panic(err)
	}
	buff := bufferpool.GetBuffer()
	doa.MustOK(tmpl.Execute(buff, t))
	res := buff.String()
	bufferpool.PutBuffer(buff)
	return res
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

func (p *PersonInfoNode) WriteAllTags(outputDir string) {
	tableStr := `# 标签表
| # | 标签 | 页面 | 力扣链接 | 解答数目 |
|:--:|:-----|:---------:|:----:|:----:|
`
	sb := strings.Builder{}
	sb.WriteString(tableStr)
	i := 1

	tagLinksSlice := make([][]TagsLink, 0, len(p.TagsMap))
	for _, tagLink := range p.TagsMap {
		tagLinksSlice = append(tagLinksSlice, tagLink)
	}
	sort.Slice(tagLinksSlice, func(i, j int) bool {
		return len(tagLinksSlice[i]) > len(tagLinksSlice[j])
	})

	for _, tagLinks := range tagLinksSlice {
		tagSlug := tagLinks[0].topicTag.Slug
		tagCn := tagLinks[0].topicTag.TranslatedName
		lkLink := fmt.Sprintf("[%s](%s%s)", tagSlug, lccli.UrlTag, tagSlug)
		localLink := fmt.Sprintf("[🔗](tags/%s.md)", tagSlug)
		tmp := fmt.Sprintf("|%d|%s|%s|%s|%d|", i, tagCn, localLink, lkLink, len(tagLinks))
		sb.WriteString(tmp)
		sb.WriteString("\n")
		i++
	}

	sb.WriteString("# 标签\n")
	for _, tagLinks := range tagLinksSlice {
		tagSlug := tagLinks[0].topicTag.Slug
		sb.WriteString(NewTagFormatter(tagSlug, tagLinks).String())
	}
	sb.WriteString("\n")

	fileName := path.Join(outputDir, "TAGS.md")
	os.WriteFile(fileName, []byte(sb.String()), os.ModePerm)
}

func (p *PersonInfoNode) OutputTags(outputDir string) error {
	mkdir(outputDir)
	for slug, tagLinks := range p.TagsMap {
		NewTagFormatter(slug, tagLinks).outPutTagMarkDown(outputDir)
	}
	return nil
}
