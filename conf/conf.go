package conf

import (
	"encoding/json"
	"io/ioutil"
)

type RootConf struct {
	ChainsConfMap map[string]string
	Consumers []string
}

type FiscoConf struct {
	SDKFile string
	EthCommon *EthLikeConf
	ChainId uint64
	StartHeight uint64
}

type EthLikeConf struct {
	Name string
	URL string
	Contracts map[string]string
	EventName map[string][]string
}

type FabConf struct {
	Name string
	EventKeyWord []string
	SDKConf string
	MSPPath string
	Channel string
	User string
	Org string
	ChainId uint64
	StartHeight uint64
}

type PolyConf struct {
	URL string
	StartHeight uint32
}

func LoadConf(ins interface{}, file string) error {
	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(raw, ins); err != nil {
		return err
	}

	return nil
}