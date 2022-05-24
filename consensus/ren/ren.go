package clique

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const BlockBuffer = 100
const LogDarknodeRegistered = "0x7c56cb7f63b6922d24414bf7c2b2c40c7ea1ea637c3f400efa766a85ecf2f093"
const LogDarknodeDeregistered = "0xf73268ea792d9dbf3e21a95ec9711f0b535c5f6c99f6b4f54f6766838086b842"
const LogNewEpoch = "0xaf2fc4796f2932ce294c3684deffe5098d3ef65dc2dd64efa80ef94eed88b01e"

func (s *Snapshot) Watch(ctx context.Context) {
	client, err := ethclient.DialContext(ctx, s.config.API)
	if err != nil {
		panic(err)
	}
	for {
		startBlockNumber := new(big.Int).SetUint64(s.LatestBlockNumber)
		lastBlockNumber := new(big.Int).Add(startBlockNumber, big.NewInt(1000))

		latestBlock, err := client.BlockByNumber(ctx, nil)
		if err != nil {
			panic(err)
		}
		if latestBlock.Number().Cmp(lastBlockNumber) < 0 {
			lastBlockNumber = latestBlock.Number()
		}

		logs, err := client.FilterLogs(ctx, ethereum.FilterQuery{
			FromBlock: startBlockNumber,
			ToBlock:   lastBlockNumber,
			Addresses: []common.Address{s.config.DNR},
			Topics: [][]common.Hash{
				{common.HexToHash(LogDarknodeRegistered)},
				{common.HexToHash(LogDarknodeDeregistered)},
				{common.HexToHash(LogNewEpoch)},
			},
		})
		if err != nil {
			panic(err)
		}

		for _, log := range logs {
			switch log.Topics[0].Hex() {
			case LogDarknodeRegistered:
				s.register(common.BytesToAddress(log.Topics[2].Bytes()))
			case LogDarknodeDeregistered:
				s.deregister(common.BytesToAddress(log.Topics[2].Bytes()))
			case LogNewEpoch:
				s.renepoch(log.BlockNumber)
			}
		}

		s.LatestBlockNumber = lastBlockNumber.Uint64()
		// if latestBlock.Number().Cmp(lastBlockNumber) == 0 {
		// 	time.Sleep(time.Hour)
		// }
	}
}
