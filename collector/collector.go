package collector

import (
	"errors"
	"github.com/realzhangm/leetcode_aid/collector/leetcode_cli"
)

var (
	ErrClient    = leetcode_cli.ErrorClient
	ErrExtractor = errors.New("extract error")
)

type Collector struct {
	ltClit    *leetcode_cli.Client
	outPutDir string
	dataDir   string
}

func NewCollector(c *Config) *Collector {
	collector := &Collector{
		ltClit:    leetcode_cli.NewClient(&c.ltClientConf),
		outPutDir: c.OutPutDir,
		dataDir:   c.DataDir,
	}
	return collector
}

func (c *Collector) getAllAcProblems() {

}

func (c *Collector) GetAllProblems() {

}

func (c *Collector) getEverySubmissionDetail() {

}

func (c *Collector) GetAllSubmissions() {

}

func (c *Collector) ExtractOneMarkDown() {

}
