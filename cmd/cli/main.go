package main

import (
	"fmt"
	flag "github.com/spf13/pflag"
	"log"
	"mc-whitelist-team-manager-cli/internal/common"
	"mc-whitelist-team-manager-cli/internal/data"
)

const (
	Version   = "0.2.1"
	Copyright = "Copyright © 2022 yzy613. All rights reserved.\n" +
		"GitHub: https://github.com/yzy613"
)

var (
	Build                = ""
	versionOption        = flag.BoolP("version", "v", false, "打印版本信息并退出")
	initOption           = flag.BoolP("init", "i", false, "初始化配置文件并退出")
	generateDataOption   = flag.StringP("generate", "g", "", "指定队伍目录并生成数据文件并退出")
	dataFile             = flag.StringP("data", "d", "", "手动指定数据文件名")
	insecure             = flag.BoolP("insecure", "k", false, "使用 SSL 时允许不安全的服务器连接")
	tpTeam               = flag.StringP("tp-team", "t", "", "指定要 tp 的队伍")
	tpCountPerCoordinate = flag.IntP("tp-count-per-coordinate", "", 1, "每个坐标传送几个玩家")
	coordinateFile       = flag.StringP("coordinate-file", "c", "", "导入每行一个坐标，每个坐标的xyz轴用空格分隔的文本文件")
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

	// 处理配置文件
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
	if *dataFile != "" {
		c.DefaultDataFileName = *dataFile
	}

	// 生成可被解析的数据
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
		err = common.MarshalAndSave(f.Teams, c.DefaultDataFileName)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("数据生成完成")
		return
	}

	// 加载可解析的数据
	err = f.LoadJsonFile(c.DefaultDataFileName)
	if err != nil {
		log.Fatalln(err)
	}
	if *tpTeam != "" {
		// 执行 tp 命令
		err = f.ParseCoordinate(*coordinateFile)
		if err != nil {
			log.Fatalln(err)
		}
		err = f.ExecuteTpCommand(c, *tpTeam, *tpCountPerCoordinate)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		// 执行白名单和分队命令
		err = f.ExecuteWhiteTeamCommand(c)
		if err != nil {
			log.Fatalln(err)
		}
	}
	log.Println("执行完成")
}
