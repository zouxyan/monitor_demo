package scanners

import (
	"github.com/ontio/monitor_demo/conf"
	"github.com/ontio/monitor_demo/core"
	"github.com/ontio/monitor_demo/log"
	"time"
)

type FabricScanner struct {
	Eng    *core.FabricEngine
	output chan *core.EventsPkg
}

func NewFabricScanner(o chan *core.EventsPkg, conf *conf.FabConf) (*FabricScanner, error) {
	eng, err := core.NewFabricEngine(conf)
	if err != nil {
		return nil, err
	}

	return &FabricScanner{
		Eng:    eng,
		output: o,
	}, nil
}

func (scanner *FabricScanner) Do() {
	info, err := scanner.Eng.LedgerCLi.QueryInfo()
	if err != nil {
		panic(err)
	}
	curr := info.BCI.Height - 1
	left := curr - 1
	if scanner.Eng.EngConf.StartHeight < left && scanner.Eng.EngConf.StartHeight != 0{
		left = scanner.Eng.EngConf.StartHeight
	}

	log.Infof("Fabric (URL: %s) scanner start from %d", "no", left)

	updateTicker := time.NewTicker(time.Second)
	for {
		select {
		case <-updateTicker.C:
			info, err := scanner.Eng.LedgerCLi.QueryInfo()
			if err != nil {
				panic(err)
			}
			curr = info.BCI.Height - 1
			if curr <= left {
				continue
			}

			for h := left + 1; h <= curr; h++ {
				blk, err := scanner.Eng.LedgerCLi.QueryBlock(h)
				if err != nil {
					log.Errorf("failed to get fabric block: %v", err)
					h--
					time.Sleep(time.Second)
					continue
				}
				pkgs, err := scanner.Eng.EventFilter(blk.Data.Data)
				if err != nil {
					log.Errorf("failed to get fabric block: %v", err)
					h--
					time.Sleep(time.Second)
					continue
				}
				for _, v := range pkgs {
					scanner.output <- v
				}
			}
		}
	}
}
