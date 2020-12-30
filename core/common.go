package core

const ZeroChainId = 0

type ChainType int

func (ty ChainType) String() string {
	switch ty {
	case EthereumTy:
		return "ethereum"
	case FiscoTy:
		return "fisco"
	case FabricTy:
		return "fabric"
	case PolyTy:
		return "poly"
	default:
		return "unsupport"
	}
}

const (
	PolyTy  ChainType = iota
	EthereumTy
	FiscoTy
	FabricTy
)

type EventsPkg struct {
	Type ChainType
	ChainId uint64
	Contract string
	TxHash string
	EventsInATx map[string]map[string]interface{}
}