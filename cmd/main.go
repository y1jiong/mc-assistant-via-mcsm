package main

import (
	"fmt"
	flag "github.com/spf13/pflag"
	"log"
	"mc-assistant-via-mcsm/internal/common"
	"mc-assistant-via-mcsm/internal/data"
	"mc-assistant-via-mcsm/internal/service"
)

const (
	Version   = "0.3.2"
	Copyright = "Copyright © 2022-2024 yzy613. All rights reserved.\n" +
		"GitHub: https://github.com/yzy613"
)

var (
	BuildTime            = ""
	CommitHash           = ""
	versionOption        = flag.BoolP("version", "v", false, "打印版本信息并退出")
	initOption           = flag.BoolP("init", "i", false, "初始化配置文件并退出")
	generateDataOption   = flag.StringP("generate", "g", "", "指定队伍目录并生成数据文件并退出")
	dataFile             = flag.StringP("data", "d", "", "手动指定数据文件名")
	insecure             = flag.BoolP("insecure", "k", false, "使用 https 链接时不检查 TLS 证书合法性")
	tpTeam               = flag.StringP("tp-team", "t", "", "指定要 tp 的队伍")
	tpCountPerCoordinate = flag.IntP("tp-count-per-coordinate", "", 1, "每个坐标传送几个玩家")
	coordinateFile       = flag.StringP("coordinate-file", "c", "", "导入每行一个坐标，每个坐标的xyz轴用空格分隔的文本文件")
	noTeam               = flag.BoolP("no-team", "N", false, "仅加白名单，不分配队伍")
	tickerInGameDay      = flag.IntP("ticker", "T", 0, "指定游戏内一天多少分钟")
	delay                = flag.IntP("delay", "D", 0, "指定每次发送命令的延迟，单位毫秒。只能大于 550 毫秒")
)

func main() {
	flag.Parse()
	if *versionOption {
		printVersion()
		return
	}

	var err error

	// 处理配置文件
	c := common.Config{}
	// 初始化配置文件
	if *initOption {
		err = c.InitToFile()
		if err != nil {
			log.Fatalln(err)
		}
		return
	}
	// 加载配置文件
	err = c.LoadFromFile()
	if err != nil {
		log.Fatalln(err)
	}
	if *dataFile != "" {
		c.DefaultDataFile = *dataFile
	}
	c.Init(*insecure)
	if *delay > c.DelayMilliseconds {
		c.SetDelay(*delay)
	}

	// mc time ticker
	if *tickerInGameDay != 0 {
		err = service.RunTicker(c, *tickerInGameDay)
		if err != nil {
			log.Fatalln(err)
		}
		return
	}

	// 生成可被解析的数据
	generateFromDirectory := *generateDataOption
	teams := data.Teams{}
	teams.Init()
	if generateFromDirectory != "" {
		if l := len(generateFromDirectory); generateFromDirectory[l-1:] == "/" || generateFromDirectory[l-1:] == "\\" {
			generateFromDirectory = generateFromDirectory[0 : l-1]
		}
		err = teams.ParseTeamAndMember(generateFromDirectory)
		if err != nil {
			log.Fatalln(err)
		}
		err = common.MarshalAndSave(teams.Teams, c.DefaultDataFile)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("数据生成完成")
		return
	}

	// 加载可解析的数据
	err = teams.LoadJsonFile(c.DefaultDataFile)
	if err != nil {
		log.Fatalln(err)
	}
	if *noTeam {
		teams.NoTeam = *noTeam
	}
	if *tpTeam != "" {
		// 执行 tp 命令
		err = teams.ParseCoordinate(*coordinateFile)
		if err != nil {
			log.Fatalln(err)
		}
		err = teams.ExecuteTpCommand(c, *tpTeam, *tpCountPerCoordinate)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		// 执行白名单和分队命令
		err = teams.ExecuteWhiteTeamCommand(c)
		if err != nil {
			log.Fatalln(err)
		}
	}
	log.Println("执行完成")
}

func printVersion() {
	fmt.Println("Version: " + Version)
	fmt.Println("Build Time: " + BuildTime)
	fmt.Println("Commit Hash: " + CommitHash)
	fmt.Println(Copyright)
}
