package ctx

import (
	"fmt"
	"github.com/ontio/monitor_demo/conf"
	"github.com/ontio/monitor_demo/consumers"
	"github.com/ontio/monitor_demo/core"
	"github.com/ontio/monitor_demo/log"
	"github.com/ontio/monitor_demo/scanners"
	"strings"
)

var (
	Ctx *Context
)

type Context struct {
	Scanners []Doer
	Bus chan *core.EventsPkg
	Consumers []Doer
}

func InitCtx(file string) {
	Ctx = &Context{
		Scanners: make([]Doer, 0),
		Bus: make(chan *core.EventsPkg),
		Consumers: make([]Doer, 0),
	}
	root := &conf.RootConf{}
	if err := conf.LoadConf(root, file); err != nil {
		panic(err)
	}
	for k, v := range root.ChainsConfMap {
		switch k {
		case "fisco":
			c := &conf.FiscoConf{}
			if err := conf.LoadConf(c, v); err != nil {
				panic(err)
			}
			for k, v := range c.EthCommon.EventName {
				delete(c.EthCommon.EventName, k)
				c.EthCommon.EventName[strings.ToLower(k)] = v
			}
			for k, v := range c.EthCommon.Contracts {
				delete(c.EthCommon.Contracts, k)
				c.EthCommon.Contracts[strings.ToLower(k)] = v
			}
			s, err := scanners.NewFiscoScanner(Ctx.Bus, c)
			if err != nil {
				panic(err)
			}
			Ctx.Scanners = append(Ctx.Scanners, s)

		case "fabric":
			c := &conf.FabConf{}
			if err := conf.LoadConf(c, v); err != nil {
				panic(err)
			}
			s, err := scanners.NewFabricScanner(Ctx.Bus, c)
			if err != nil {
				panic(err)
			}
			Ctx.Scanners = append(Ctx.Scanners, s)

		case "ethereum":

		case "poly":
			c := &conf.PolyConf{}
			if err := conf.LoadConf(c, v); err != nil {
				panic(err)
			}
			s, err := scanners.NewPolyScanner(Ctx.Bus, c)
			if err != nil {
				panic(err)
			}
			Ctx.Scanners = append(Ctx.Scanners, s)

		default:
			panic(fmt.Errorf("config %s for %s not supported", v, k))
		}
	}

	for _, ty := range root.Consumers {
		switch ty {
		case "core":
			processor := consumers.NewCrossStatusProcessor(Ctx.Bus)
			Ctx.Consumers = append(Ctx.Consumers, processor)
		default:
			panic(fmt.Errorf("%s consumer not supported", ty))
		}
	}
}

func (ctx *Context) Run() {
	log.Info("start all doers")
	for _, c := range ctx.Consumers {
		go c.Do()
	}

	for _, s := range ctx.Scanners {
		go s.Do()
	}
}