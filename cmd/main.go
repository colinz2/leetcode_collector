package main

import "github.com/realzhangm/leetcode_aid/collector"

// 实现简单的功能
// 将 Leetcode 上的提交的代码，同步到本地，并生产一个 Markdown 汇总文件。

func extractOneMarkDown() {
	c := collector.NewCollector(collector.GetConfig())
	if err := c.LoadInfo(); err != nil{
		panic(err)
	}

	err := c.FetchFromLeetCode()
	if err != nil {
		panic(err)
	}

	err = c.ExtractOneMarkDown()
	if err != nil {
		panic(err)
	}
}

func json2Md() {
	c := collector.NewCollector(collector.GetConfig())
	if err := c.LoadInfo(); err != nil{
		panic(err)
	}
	c.Json2MD()
	c.OutputSolutions()
}

func main() {
	extractOneMarkDown()
	json2Md()
}
