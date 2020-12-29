package consumers

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/monitor_demo/log"
	"github.com/ontio/monitor_demo/scanners"
	"github.com/polynetwork/poly/common"
	"time"
)

type StagePosition int

const (
	FromStage StagePosition = iota
	PolyStage
	ToStage
)

type Stage struct {
	ChainTy scanners.ChainType
	ChainId uint64
	TxHash string
}

type Status struct {
	From, To uint64
	TxId []byte
	Stages []*Stage
}

func (s *Status) IsDone() bool {
	return len(s.Stages) == 3 && s.Stages[FromStage] != nil && s.Stages[PolyStage] != nil && s.Stages[ToStage] != nil
}

func (s *Status) AddStage(stage *Stage, pos StagePosition) {
	if len(s.Stages) > int(pos) {
		s.Stages[pos] = stage
		return
	}
	s.Stages = append(s.Stages, stage)
}

func (s *Status) Show() string {
	from := "not received"
	if sg := s.Stages[FromStage]; sg != nil {
		from = fmt.Sprintf("(chain_type: %s, chain_id: %d, txhash: %s)", sg.ChainTy, sg.ChainId, sg.TxHash)
	}
	to := "not received"
	if sg := s.Stages[ToStage]; sg != nil {
		to = fmt.Sprintf("(chain_type: %s, chain_id: %d, txhash: %s)", sg.ChainTy, sg.ChainId, sg.TxHash)
	}
	poly := "not received"
	if sg := s.Stages[PolyStage]; sg != nil {
		poly = fmt.Sprintf("(txhash: %s)", sg.TxHash)
	}

	return fmt.Sprintf(
		"[ txid: %s, from: %s, poly: %s, to: %s ]",
		hex.EncodeToString(s.TxId), from, poly, to,
		)
}

type CrossStatusProcessor struct {
	recev chan <-*scanners.EventsPkg
	data map[string]*Status
}

func (p *CrossStatusProcessor) Receiving() {
	for item := range p.recev {
		p.getEventStage(item)
	}
}

func (p *CrossStatusProcessor) getEventStage(item *scanners.EventsPkg) {
	switch item.Type {
	case scanners.FiscoTy:
		p.handleFiscoEvent(item)
	case scanners.PolyTy:
		p.handlePolyEvent(item)
	case scanners.FabricTy:

	}
}

func (p *CrossStatusProcessor) handleFiscoEvent(item *scanners.EventsPkg) {
	e, ok := item.EventsInATx["CrossChainEvent"]
	if ok {
		txId := e["txId"].([]byte)
		toChainId := e["toChainId"].(uint64)
		k := makeCCKey(txId, item.ChainId, toChainId)
		v, ok := p.data[k]
		stage := &Stage{
			ChainTy: scanners.FiscoTy,
			ChainId: item.ChainId,
			TxHash: item.TxHash,
		}
		if !ok {
			p.data[k] = &Status{
				From: item.ChainId,
				To: toChainId,
				TxId: txId,
				Stages: []*Stage{
					stage,
				},
			}
		} else {
			v.AddStage(stage, FromStage)
		}
	}

	e, ok = item.EventsInATx["VerifyHeaderAndExecuteTxEvent"]
	if ok {
		txId := e["fromChainTxHash"].([]byte)
		fromChainID := e["fromChainID"].(uint64)
		k := makeCCKey(txId, fromChainID, item.ChainId)
		v, ok := p.data[k]
		stage := &Stage{
			ChainTy: scanners.FiscoTy,
			ChainId: item.ChainId,
			TxHash: item.TxHash,
		}
		if !ok {
			p.data[k] = &Status{
				From: fromChainID,
				To: item.ChainId,
				TxId: txId,
				Stages: []*Stage{
					nil,
					nil,
					stage,
				},
			}
		} else {
			v.AddStage(stage, ToStage)
		}
	}
}

func (p *CrossStatusProcessor) handlePolyEvent(item *scanners.EventsPkg) {
	for _, v := range item.EventsInATx {
		txId := v["txid"].(string)
		fromChainId := v["from"].(uint64)
		toChainId := v["to"].(uint64)
		rawTxId, _ := hex.DecodeString(txId)
		k := makeCCKey(rawTxId, fromChainId, toChainId)

		v, ok := p.data[k]
		stage := &Stage{
			ChainTy: scanners.PolyTy,
			ChainId: item.ChainId,
			TxHash: item.TxHash,
		}
		if !ok {
			p.data[k] = &Status{
				From: fromChainId,
				To: toChainId,
				TxId: rawTxId,
				Stages: []*Stage{
					nil,
					stage,
					nil,
				},
			}
		} else {
			v.AddStage(stage, PolyStage)
		}
	}
}

func makeCCKey(txid []byte, from, to uint64) string {
	sink := common.NewZeroCopySink(nil)
	sink.WriteVarBytes(txid)
	sink.WriteUint64(from)
	sink.WriteUint64(to)

	return hex.EncodeToString(sink.Bytes())
}

func (p *CrossStatusProcessor) showStatus() string {
	res := "{\n"
	for i, v := range p.data {
		res += fmt.Sprintf("\tNo.%d, info: %s\n", i, v.Show())
	}
	return res + "}\n"
}

func (p *CrossStatusProcessor) PrintStatus() {
	tick := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-tick.C:
			log.Info(p.showStatus())
		}
	}
}