package scanners

import (
	"context"
	"github.com/polynetwork/monitor_demo/conf"
	"github.com/polynetwork/monitor_demo/core"
	"github.com/polynetwork/monitor_demo/log"
	"math/big"
	"time"
)

type EthScanner struct {
	Eng *core.EthLikeEngine
	config *conf.EthereumConf
	output chan <-*core.EventsPkg
}

func NewEthScanner(o chan <-*core.EventsPkg, conf *conf.EthereumConf) (*EthScanner, error) {
	eng, err := core.NewEthLikeEngine(conf.EthCommon)
	if err != nil {
		panic(err)
	}

	return &EthScanner{
		Eng: eng,
		config: conf,
		output: o,
	}, nil
}

func (scanner *EthScanner) Do() {
	left, err := scanner.GetBlockNumber()
	if err != nil {
		panic(err)
	}
	if scanner.config.StartHeight != 0 && scanner.config.StartHeight < left {
		left = scanner.config.StartHeight
	}

	updateTicker := time.NewTicker(time.Second)
	for {
		select {
		case <-updateTicker.C:
			right, err := scanner.GetBlockNumber()
			if err != nil {
				log.Errorf("EthScanner failed to BlockNumber: %v", err)
				continue
			}
			if right < left + scanner.config.BlocksToWait {
				continue
			}
			for left + scanner.config.BlocksToWait <= right {
				err := scanner.check(left)
				if err != nil {
					log.Errorf("EthScanner error: %v", err)
					continue
				}
				left++
			}
		}
	}
}

func (scanner *EthScanner) check(h uint64) error {
	blk, err := scanner.Eng.Cli.BlockByNumber(context.Background(), big.NewInt(0).SetUint64(h))
	if err != nil {
		return err
	}
	txns := blk.Transactions()
	for _, tx := range txns {
		recp, err := scanner.Eng.Cli.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			if err != nil {
				log.Errorf("CheckEthHeight - tx %s TransactionReceipt error: %v", tx, err.Error())
				continue
			}
			if recp.Status != 1 {
				continue
			}
			for _, v := range recp.Logs {
				_, logs, err := scanner.Eng.EventFilter(*v)
				if len(logs) == 0 || err != nil {
					continue
				}

				scanner.output <- &core.EventsPkg {
					Type: core.FiscoTy,
					Contract: v.Address.String(),
					EventsInATx: logs,
					ChainId: scanner.config.ChainId,
					TxHash: tx.Hash().String(),
				}
			}
		}
	}

	return nil
}

func (scanner *EthScanner) GetBlockNumber() (uint64, error) {
	hdr, err := scanner.Eng.Cli.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	return hdr.Number.Uint64(), nil
}