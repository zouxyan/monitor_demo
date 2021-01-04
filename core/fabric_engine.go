package core

import (
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/polynetwork/monitor_demo/conf"
	"github.com/polynetwork/monitor_demo/core/internal/github.com/hyperledger/fabric/protoutil"
	"os"
	"strings"
)

type FabricEngine struct {
	Sdk       *fabsdk.FabricSDK
	ChanCli   *channel.Client
	EventCli  *event.Client
	LedgerCLi *ledger.Client
	EngConf *conf.FabConf
}

func NewFabricEngine(conf *conf.FabConf) (*FabricEngine, error) {
	_ = os.Setenv("FABRIC_MSP_PATH" + conf.Name, conf.MSPPath)

	sdk, err := fabsdk.New(config.FromFile(conf.SDKConf))
	if err != nil {
		return nil, err
	}
	ccp := sdk.ChannelContext(conf.Channel, fabsdk.WithUser(conf.User), fabsdk.WithOrg(conf.Org))
	cc, err := channel.New(ccp)
	if err != nil {
		return nil, err
	}
	eventClient, err := event.New(ccp, event.WithBlockEvents())
	if err != nil {
		return nil, err
	}
	ledgerClient, err := ledger.New(ccp)
	if err != nil {
		return nil, err
	}

	return &FabricEngine{
		Sdk: sdk,
		LedgerCLi: ledgerClient,
		ChanCli: cc,
		EventCli: eventClient,
		EngConf: conf,
	}, nil
}

func (eng *FabricEngine) EventFilter(data [][]byte) ([]*EventsPkg, error) {
	res := make([]*EventsPkg, 0)
	for _, v := range data {
		xx, _ := protoutil.GetEnvelopeFromBlock(v)
		cas, _ := protoutil.GetActionsFromEnvelopeMsg(xx)

		if len(cas) == 0 {
			continue
		}

		pkg := &EventsPkg{
			Contract: cas[0].ChaincodeId.Name,
			Type:     FabricTy,
			ChainId:  eng.EngConf.ChainId,
			EventsInATx: make(map[string]map[string]interface{}),
		}
		for _, v := range cas {
			chaincodeEvent := &peer.ChaincodeEvent{}
			_ = proto.Unmarshal(v.Events, chaincodeEvent)
			if len(chaincodeEvent.TxId) == 0 {
				continue
			}

			for _, w := range eng.EngConf.EventKeyWord {
				if strings.Contains(chaincodeEvent.EventName, w) {
					tx, _ := eng.LedgerCLi.QueryTransaction(fab.TransactionID(chaincodeEvent.TxId))
					if tx.ValidationCode != 0 {
						continue
					}

					pkg.TxHash = chaincodeEvent.TxId
					pkg.EventsInATx[w] = map[string]interface{}{
						"raw": chaincodeEvent.Payload,
					}

					break
				}
			}
		}

		if len(pkg.EventsInATx) > 0 {
			res = append(res, pkg)
		}
	}

	return res, nil
}