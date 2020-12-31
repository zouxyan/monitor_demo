package scanners

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/FISCO-BCOS/go-sdk/client"
	conf2 "github.com/FISCO-BCOS/go-sdk/conf"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ontio/monitor_demo/conf"
	"github.com/ontio/monitor_demo/core"
	"github.com/ontio/monitor_demo/log"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type FiscoScanner struct {
	Eng *core.EthLikeEngine
	SDK *client.Client
	chainId uint64
	currh uint64
	output chan <-*core.EventsPkg
}

func NewFiscoScanner(o chan <-*core.EventsPkg, conf *conf.FiscoConf) (*FiscoScanner, error) {
	eng, err := core.NewEthLikeEngine(conf.EthCommon)
	if err != nil {
		return nil, err
	}
	cset := conf2.ParseConfig(conf.SDKFile)
	if len(cset) == 0 {
		return nil, errors.New("no sdk config")
	}
	cli, err := client.Dial(&cset[0])
	if err != nil {
		return nil, err
	}

	return &FiscoScanner{
		Eng: eng,
		SDK: cli,
		currh: conf.StartHeight,
		output: o,
		chainId: conf.ChainId,
	}, nil
}

func (scanner *FiscoScanner) Do() {
	left, err := scanner.BlockNumber()
	if err != nil {
		panic(err)
	}
	left--
	if left > scanner.currh && scanner.currh != 0{
		left = scanner.currh
	}
	log.Infof("Fisco (URL: %s) scanner start from %d", scanner.Eng.EngConf.URL, left)

	updateTicker := time.NewTicker(time.Second)
	for {
		select {
		case <-updateTicker.C:
			right, err := scanner.BlockNumber()
			if err != nil {
				log.Errorf("MonitorFisco failed to BlockNumber: %v", err)
				continue
			}
			if right <= left {
				continue
			}
			for left < right {
				left++
				err := scanner.check(left)
				if err != nil {
					log.Errorf("MonitorFisco error: %v", err)
				}
			}
		}
	}
}

func (scanner *FiscoScanner) check(height uint64) error {
	blk, err := scanner.SDK.GetBlockByNumber(context.Background(), strconv.FormatUint(height, 10), false)
	if err != nil {
		return fmt.Errorf("CheckFiscoHeight - GetBlockByNumber error :%s", err.Error())
	}
	res := &BlockRes{}
	err = json.Unmarshal(blk, res)
	if err != nil {
		return fmt.Errorf("CheckFiscoHeight - Unmarshal error :%s", err.Error())
	}

	for _, tx := range res.Transactions {
		recp, err := scanner.SDK.TransactionReceipt(context.Background(), common.HexToHash(tx))
		if err != nil {
			log.Errorf("CheckFiscoHeight - tx %s TransactionReceipt error: %v", tx, err.Error())
			continue
		}
		if recp.Status != "0x0" {
			continue
		}
		for _, v := range recp.Logs {
			topics := make([]common.Hash, len(v.Topics))
			for i, t := range v.Topics {
				topics[i] = common.HexToHash(t.(string))
			}
			rawData, _ := hex.DecodeString(strings.TrimPrefix(v.Data, "0x"))
			_, logs, err := scanner.Eng.EventFilter(types.Log{
				Address: common.HexToAddress(v.Address),
				Topics:  topics,
				Data:    rawData,
			})
			if len(logs) == 0 || err != nil {
				continue
			}

			scanner.output <- &core.EventsPkg {
				Type: core.FiscoTy,
				Contract: v.Address,
				EventsInATx: logs,
				ChainId: scanner.chainId,
				TxHash: tx,
			}
		}
	}

	return nil
}

func (scanner *FiscoScanner) BlockNumber() (uint64, error) {
	bn, err := scanner.SDK.GetBlockNumber(context.Background())
	if err != nil {
		return 0, fmt.Errorf("block number not found: %v", err)
	}
	str, err := strconv.Unquote(*(*string)(unsafe.Pointer(&bn)))
	if err != nil {
		return 0, fmt.Errorf("ParseInt: %v", err)
	}
	height, err := strconv.ParseUint(str, 0, 0)
	if err != nil {
		return 0, fmt.Errorf("ParseInt: %v", err)
	}
	return height, nil
}

type BlockRes struct {
	Transactions []string `json:"transactions"`
}