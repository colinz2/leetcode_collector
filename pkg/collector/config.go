package collector

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
)

import (
	"github.com/howeyc/gopass"
	"github.com/jinzhu/configor"
	"github.com/pkg/errors"
	"github.com/realzhangm/leetcode_collector/pkg/collector/leetcode_cli"
)

type Config struct {
	ltClientConf leetcode_cli.ClientConf
	SolutionsDir string
	OutputDir    string `default:"./output"`
	initFlag     bool
}

var (
	config Config
)

func GetConfig() *Config {
	if !config.initFlag {
		panic("not init")
	}
	return &config
}

func init() {
	LoadConfig()
}

func LoadConfig() {
	// file not exist, no error return
	err := configor.Load(&config, "config.toml")
	if err != nil {
		fmt.Println("---------------")
		panic(err)
	}

	if config.OutputDir == "" {
		config.OutputDir = "./output"
	}

	if config.SolutionsDir == "" {
		config.SolutionsDir = path.Join(config.OutputDir, "solutions")
	}

	if loadPass() != nil {
		var err error
		config.ltClientConf.UserName, config.ltClientConf.PassWord, err = credentials()
		if err != nil {
			panic(err)
		}
		savePass()
	}

	config.initFlag = true
}

func loadPass() error {
	buf, err := os.ReadFile(".password")
	if err != nil {
		return err
	}
	sn := strings.SplitN(string(buf), " ", 2)
	if len(sn) != 2 {
		return errors.New("format not right")
	}
	config.ltClientConf.UserName = sn[0]
	config.ltClientConf.PassWord = sn[1]
	return nil
}

func savePass() {
	str := fmt.Sprintf("%s %s", config.ltClientConf.UserName, config.ltClientConf.PassWord)
	err := os.WriteFile(".password", []byte(str), os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func credentials() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	fmt.Print("Enter Password: ")
	pass, err := gopass.GetPasswd()
	if err != nil {
		return "", "", err
	}

	password := string(pass)
	return strings.TrimSpace(username), strings.TrimSpace(password), nil
}
