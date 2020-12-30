package main

import (
	"flag"
	"github.com/ontio/monitor_demo/ctx"
)

var (
	confPath  string
)

func init() {
	flag.StringVar(&confPath, "conf", "./conf.json", "configuration file for main process")
}

func main() {
	flag.Parse()

	ctx.InitCtx(confPath)
	ctx.Ctx.Run()

	select {}
}

