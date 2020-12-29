package core

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ontio/monitor_demo/conf"
	"github.com/ontio/monitor_demo/log"
	"strings"
)

type EthLikeEngine struct {
	Cli *ethclient.Client
	ABIs map[string]abi.ABI
	EngConf *conf.EthLikeConf
}

func NewEthLikeEngine(conf *conf.EthLikeConf) (*EthLikeEngine, error) {
	cli, err := ethclient.Dial(conf.URL)
	if err != nil {
		return nil, err
	}
	abis := make(map[string]abi.ABI)
	for k, v := range conf.Contracts {
		abiIns, err := GetABI(v)
		if err != nil {
			return nil, fmt.Errorf("failed to get abi for contract %s: %v", k, err)
		}
		abis[k] = abiIns
	}

	return &EthLikeEngine {
		Cli: cli,
		ABIs: abis,
		EngConf: conf,
	}, nil
}

func (e *EthLikeEngine) EventFilter(l types.Log) (string, map[string]map[string]interface{}, error) {
	res := make(map[string]map[string]interface{}, 0)
	contract := strings.ToLower(l.Address.String())
	words, ok := e.EngConf.EventName[contract]
	if !ok {
		return "", nil, nil
	}
	a, ok := e.ABIs[contract]
	if !ok {
		return "", nil, fmt.Errorf("contract %s not found in ABIs", contract)
	}
	for _, v := range words {
		_, ok = a.Events[v]
		if !ok {
			log.Warnf("event %s not found in contract %s", v, contract)
			continue
		}
		event := make(map[string]interface{})
		if err := UnpackLogIntoMap(a, event, v, l); err != nil {
			return "", nil, err
		}
		res[v] = event
	}
	return contract, res, nil
}