package main

import (
	"flag"
	"fmt"

	"os"

	"strings"

	"github.com/astaxie/beego"
	"github.com/fatih/color"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	cfg = viper.New()
)

func initConfig() (err error) {
	cfg.SetDefault("agentX.version", "1.0")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	//cli&default config
	configFile := pflag.String("config", "", "config file path")
	version := pflag.Bool("version", false, "show version")
	if *version {
		fmt.Printf("agentX v%s - https://github.com/snail007/agentX\n", cfg.GetString("agentX.version"))
		os.Exit(0)
	}
	pflag.Int("http-port", 9091, "the api port to listen")
	pflag.String("run-mode", "dev", "web service run mode,should be on of dev|prod|test")
	pflag.String("level", "debug", "log level to show in console")
	pflag.String("log-dir", "log", "the directory which store log files")
	pflag.Int64("log-max-size", 102400000, "log file max size(bytes) for rotate")
	pflag.Int("log-max-count", 3, "log file max count for rotate to remain")
	pflag.StringSlice("log-level", []string{"info", "error", "debug"}, "log to file level,multiple splitted by comma(,)")
	pflag.Parse()

	//bind flag
	cfg.BindPFlag("web.HttpPort", pflag.Lookup("http-port"))
	cfg.BindPFlag("web.RunMode", pflag.Lookup("run-mode"))
	cfg.BindPFlag("log.dir", pflag.Lookup("log-dir"))
	cfg.BindPFlag("log.level", pflag.Lookup("log-level"))
	cfg.BindPFlag("log.console-level", pflag.Lookup("level"))
	cfg.BindPFlag("log.fileMaxSize", pflag.Lookup("log-max-size"))
	cfg.BindPFlag("log.maxCount", pflag.Lookup("log-max-count"))
	if *configFile != "" {
		cfg.SetConfigFile(*configFile)
	} else {
		cfg.SetConfigName("config")
		cfg.AddConfigPath("/etc/agentx/")
		cfg.AddConfigPath("$HOME/.agentx")
		cfg.AddConfigPath("conf")
		cfg.AddConfigPath(".agentx")
		cfg.AddConfigPath(".")
	}
	err = cfg.ReadInConfig()
	file := cfg.ConfigFileUsed()
	if err != nil && !strings.Contains(fmt.Sprintf("%s", err.Error()), "Not") {
		fmt.Printf("%s", err)
	} else if file != "" {
		fmt.Printf("use config file : %s\n", file)
	}
	initBeegoConfig()
	setInternalConfig()
	return
}
func initBeegoConfig() {
	beego.BConfig.AppName = cfg.GetString("appname")
	beego.BConfig.Listen.HTTPPort = cfg.GetInt("web.HttpPort")
	beego.BConfig.RunMode = cfg.GetString("web.RunMode")
}
func setInternalConfig() {
	cfg.Set("appname", "agentX")
}
func poster() string {
	fg := color.New(color.FgHiYellow).SprintFunc()
	return fg(`
 █████╗      ██████╗     ███████╗    ███╗   ██╗    ████████╗    ██╗  ██╗
██╔══██╗    ██╔════╝     ██╔════╝    ████╗  ██║    ╚══██╔══╝    ╚██╗██╔╝
███████║    ██║  ███╗    █████╗      ██╔██╗ ██║       ██║        ╚███╔╝ 
██╔══██║    ██║   ██║    ██╔══╝      ██║╚██╗██║       ██║        ██╔██╗ 
██║  ██║    ╚██████╔╝    ███████╗    ██║ ╚████║       ██║       ██╔╝ ██╗
╚═╝  ╚═╝     ╚═════╝     ╚══════╝    ╚═╝  ╚═══╝       ╚═╝       ╚═╝  ╚═╝
Author: snail
Link  : https://github.com/snail007/agentX
`)
}
