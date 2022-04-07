package main

import (
	collector "github.com/realzhangm/leetcode_collector/pkg/collector"
	"github.com/realzhangm/leetcode_collector/pkg/doa"
)

func main() {
	c := collector.NewCollector(collector.GetConfig())
	doa.MustOK(c.LoadInfo())
	doa.MustOK(c.JsonToMarkDown())
	doa.MustOK(c.OutputSolutionsCode())
	doa.MustOK(c.OutputTagsMarkDown())
}
