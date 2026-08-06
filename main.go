package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/pflag"

	"mc-assistant-via-mcsm/internal/common"
	"mc-assistant-via-mcsm/internal/data"
	"mc-assistant-via-mcsm/internal/service"
)

const (
	Version   = "0.3.4"
	Copyright = "Copyright © 2022 y1jiong. All rights reserved.\n" +
		"GitHub: https://github.com/y1jiong"
)

var (
	GitTag    string
	GitCommit string
	BuildTime string
)

type options struct {
	version              bool
	initialize           bool
	generateDirectory    string
	dataFile             string
	insecure             bool
	tpTeam               string
	tpCountPerCoordinate int
	coordinateFile       string
	noTeam               bool
	tickerInGameDay      int
	delayMilliseconds    int
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Args[1:]); err != nil &&
		!errors.Is(err, pflag.ErrHelp) &&
		!errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}

func run(ctx context.Context, args []string) error {
	opts, err := parseOptions(args)
	if err != nil {
		return err
	}
	if opts.version {
		printVersion()
		return nil
	}

	config := common.Config{}
	if opts.initialize {
		return config.InitToFile()
	}
	if err := config.LoadFromFile(); err != nil {
		return fmt.Errorf("加载配置文件: %w", err)
	}
	if opts.dataFile != "" {
		config.DefaultDataFile = opts.dataFile
	}
	config.Init(opts.insecure)
	if opts.delayMilliseconds > config.DelayMilliseconds {
		config.SetDelay(opts.delayMilliseconds)
	}

	if opts.tickerInGameDay != 0 {
		return service.RunTicker(ctx, &config, opts.tickerInGameDay)
	}

	teams := data.NewTeams()
	if opts.generateDirectory != "" {
		if err := teams.ParseTeamAndMember(filepath.Clean(opts.generateDirectory)); err != nil {
			return err
		}
		if err := common.SaveJSON(teams.Teams, config.DefaultDataFile); err != nil {
			return err
		}
		log.Println("数据生成完成")
		return nil
	}

	if err := teams.LoadJSONFile(config.DefaultDataFile); err != nil {
		return err
	}
	teams.NoTeam = opts.noTeam

	if opts.tpTeam != "" {
		if err := teams.ParseCoordinate(opts.coordinateFile); err != nil {
			return err
		}
		if err := teams.ExecuteTPCommand(ctx, &config, opts.tpTeam, opts.tpCountPerCoordinate); err != nil {
			return err
		}
	} else if err := teams.ExecuteWhiteTeamCommand(ctx, &config); err != nil {
		return err
	}

	log.Println("执行完成")
	return nil
}

func parseOptions(args []string) (options, error) {
	var opts options
	flags := pflag.NewFlagSet("mc-assistant-via-mcsm", pflag.ContinueOnError)
	flags.BoolVarP(&opts.version, "version", "V", false, "打印版本信息并退出")
	flags.BoolVarP(&opts.initialize, "init", "i", false, "初始化配置文件并退出")
	flags.StringVarP(&opts.generateDirectory, "generate", "g", "", "指定队伍目录并生成数据文件并退出")
	flags.StringVarP(&opts.dataFile, "data", "d", "", "手动指定数据文件名")
	flags.BoolVarP(&opts.insecure, "insecure", "k", false, "使用 https 链接时不检查 TLS 证书合法性")
	flags.StringVarP(&opts.tpTeam, "tp-team", "t", "", "指定要 tp 的队伍")
	flags.IntVar(&opts.tpCountPerCoordinate, "tp-count-per-coordinate", 1, "每个坐标传送几个玩家")
	flags.StringVarP(&opts.coordinateFile, "coordinate-file", "c", "", "导入每行一个坐标，每个坐标的 xyz 轴用空格分隔的文本文件")
	flags.BoolVarP(&opts.noTeam, "no-team", "N", false, "仅加白名单，不分配队伍")
	flags.IntVarP(&opts.tickerInGameDay, "ticker", "T", 0, "指定游戏内一天多少分钟")
	flags.IntVarP(&opts.delayMilliseconds, "delay", "D", 0, "指定每次发送命令的延迟，单位毫秒。只能大于 550 毫秒")

	return opts, flags.Parse(args)
}

func printVersion() {
	fmt.Println("Version: " + Version)
	fmt.Println("Git Tag: " + GitTag)
	fmt.Println("Git Commit: " + GitCommit)
	fmt.Println("Build Time: " + BuildTime)
	fmt.Println(Copyright)
}
