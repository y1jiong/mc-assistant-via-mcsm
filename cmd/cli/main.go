package main

import (
	"fmt"
	flag "github.com/spf13/pflag"
	"log"
	"mc-whitelist-team-manager-cli/internal/common"
	"mc-whitelist-team-manager-cli/internal/data"
)

const (
	Version   = "0.1.8"
	Copyright = "Copyright © 2022 yzy613. All rights reserved.\n" +
		"GitHub: https://github.com/yzy613"
)

var (
	Build              = ""
	versionOption      = flag.BoolP("version", "v", false, "打印版本信息并退出")
	initOption         = flag.BoolP("init", "i", false, "初始化配置文件并退出")
	generateDataOption = flag.StringP("generate", "g", "", "指定队伍目录并生成数据文件并退出")
	dataFileName       = flag.StringP("data", "d", "", "忽视配置文件设置的数据文件名并指定数据文件名")
	insecure           = flag.BoolP("insecure", "k", false, "使用 SSL 时允许不安全的服务器连接")
)

func main() {
	flag.Parse()
	if *versionOption {
		fmt.Println("Version " + Version)
		fmt.Println("Build " + Build)
		fmt.Println(Copyright)
		return
	}

	var err error

	c := common.Config{}
	if *initOption {
		err = c.InitToFile()
		if err != nil {
			log.Fatalln(err)
		}
		return
	}
	err = c.LoadFromFile()
	if err != nil {
		log.Fatalln(err)
	}
	if *dataFileName != "" {
		c.DataFileName = *dataFileName
	}

	generateFromDirectory := *generateDataOption
	f := data.TeamsFormat{}
	f.Init(*insecure)
	if generateFromDirectory != "" {
		if l := len(generateFromDirectory); generateFromDirectory[l-1:] == "/" || generateFromDirectory[l-1:] == "\\" {
			generateFromDirectory = generateFromDirectory[0 : l-1]
		}
		err = f.ParseTeamAndMember(generateFromDirectory)
		if err != nil {
			log.Fatalln(err)
		}
		err = common.MarshalAndSave(f.Teams, c.DataFileName)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("数据生成完成")
		return
	}

	err = f.LoadJsonFile(c.DataFileName)
	if err != nil {
		log.Fatalln(err)
	}
	err = f.ExecuteCommand(c.ApiUrl, c.ApiKey, c.ServerName)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("执行完成")
}
