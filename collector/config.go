package collector

import "github.com/realzhangm/leetcode_aid/collector/leetcode_cli"

type Config struct {
	ltClientConf leetcode_cli.ClientConf
	OutPutDir    string
	DataDir      string
}

var (
	config Config
)

func LoadConfig() {

}
