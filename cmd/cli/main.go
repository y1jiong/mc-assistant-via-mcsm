package main

import (
	"flag"
	"fmt"
	"log"
	"mc-whitelist-team-manager-cli/internal/common"
	"mc-whitelist-team-manager-cli/internal/data"
)

const (
	Version   = "0.1.5"
	Copyright = "Copyright © 2022 yzy613. All rights reserved.\n" +
		"GitHub: https://github.com/yzy613"
)

var (
	Build              = ""
	versionOption      = flag.Bool("v", false, "打印版本信息并退出")
	initOption         = flag.Bool("i", false, "初始化配置文件并退出")
	generateDataOption = flag.String("g", "", "指定队伍目录并生成数据文件并退出")
	dataFileName       = flag.String("d", "", "忽视配置文件数据文件名并指定数据文件名")
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
	f := data.FinalFormat{}
	if generateFromDirectory != "" {
		if l := len(generateFromDirectory); generateFromDirectory[l-1:] == "/" || generateFromDirectory[l-1:] == "\\" {
			generateFromDirectory = generateFromDirectory[0 : l-1]
		}
		err = f.ParseTeamAndMember(generateFromDirectory)
		if err != nil {
			log.Fatalln(err)
		}
		err = common.MarshalAndSave(f, c.DataFileName)
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
