package scanners

import (
	"github.com/ontio/monitor_demo/conf"
	"github.com/ontio/monitor_demo/core"
	"github.com/ontio/monitor_demo/log"
	"time"
)

type PolyScanner struct {
	Eng *core.PolyEngine
	output chan <-*EventsPkg
}

func NewPolyScanner(o chan <-*EventsPkg, conf *conf.PolyConf) (*PolyScanner, error) {
	eng, err := core.NewPolyEngine(conf)
	if err != nil {
		return nil, err
	}

	return &PolyScanner{
		Eng: eng,
		output: o,
	}, nil
}

func (scanner *PolyScanner) Do() {
	currh, err := scanner.Eng.Cli.GetCurrentBlockHeight()
	if err != nil {
		panic(err)
	}

	if currh > scanner.Eng.EngConf.StartHeight {
		currh = scanner.Eng.EngConf.StartHeight
	}

	updateTicker := time.NewTicker(time.Second)
	for {
		select {
		case <-updateTicker.C:
			currentHeight, err := scanner.Eng.Cli.GetCurrentBlockHeight()
			if err != nil {
				log.Errorf("GetCurrentBlockHeight error: %s", err)
				continue
			}
			if currentHeight <= currh {
				continue
			}
			for currentHeight > currh {
				currh++
				res, err := scanner.Eng.EventFilter(currh)
				if err != nil {
					log.Errorf("parseRelayChainBlock error: %s", err)
					currh--
					break
				}

				for _, v := range res {
					scanner.output <-v
				}
			}
		}
	}
}
