package core

import (
	"github.com/ontio/monitor_demo/conf"
	"github.com/polynetwork/poly-go-sdk"
)

type PolyEngine struct {
	Cli *poly_go_sdk.PolySdk
	EngConf *conf.PolyConf
}

func NewPolyEngine(conf *conf.PolyConf) (*PolyEngine, error) {
	sdk := poly_go_sdk.NewPolySdk()
	sdk.NewRpcClient().SetAddress(conf.URL)

	return &PolyEngine{
		EngConf: conf,
		Cli: sdk,
	}, nil
}

func (eng *PolyEngine) EventFilter(height uint32) ([]*EventsPkg, error) {
	events, err := eng.Cli.GetSmartContractEventByBlock(height)
	if err != nil {
		return nil, err
	}

	res := make([]*EventsPkg, 0)
	for _, event := range events {
		arr := make(map[string]map[string]interface{})
		for _, notify := range event.Notify {
			states, ok := notify.States.([]interface{})
			if !ok {
				continue
			}
			name, ok := states[0].(string)
			if ok && name == "makeProof" {
				txid := states[3].(string)
				arr[txid] = map[string]interface{}{
					"txid": txid,
					"from": states[1].(uint64),
					"to": states[2].(uint64),
				}
			}
		}

		if len(arr) > 0 {
			res = append(res, &EventsPkg{
				TxHash: event.TxHash,
				ChainId: ZeroChainId,
				Type: PolyTy,
				Contract: "",
				EventsInATx: arr,
			})
		}
	}

	return res, nil
}