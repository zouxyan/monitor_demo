package conf

const ZeroChainId = 0

var (
	RConf *RootConf
)

func InitRootConf(f string) {

}

type RootConf struct {
	EthLikeArr []*EthLikeConf
	FabArr []*FabConf
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
	EventKeyWord map[string][]string
	SDKConf string
	MSPPath string
	Channel string
	User string
	Org string
	ChainId uint64
}

type PolyConf struct {
	URL string
	StartHeight uint32
}

