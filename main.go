package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/params"
)

var (
	dbPath          = flag.String("db_dir", "geth/chaindata", "db path to give")
	whatBlockNumber = flag.Uint("blknum", 0, "at what block number eval the input")
	senderFlag      = flag.String("sender", "", "what 0x address to use as msg.sender")
	toAddrFlag      = flag.String("receiver", "",
		"what 0x address to use as receiver (prob a smart contract)",
	)
	gasLimit = flag.Uint("limit", 500_000, "gas limit")
	gasPrice = flag.Uint("gasPrice", 1e9, "gas price")
)

func program() error {

	blkNum := *whatBlockNumber
	if blkNum == 0 {
		return errors.New("block number should not be 0 - doubt you want genisis eval")
	}

	sender := common.HexToAddress(*senderFlag)
	if sender == (common.Address{}) {
		return errors.New("invalid sender")
	}

	toAddr := common.HexToAddress(*toAddrFlag)
	if sender == (common.Address{}) {
		return errors.New("invalid receiver")
	}

	payload := flag.Args()
	if len(payload) != 1 {
		return errors.New("ought to be single EVM input data to run")
	}

	inputRunning := []byte(payload[0])
	chainDb, err := rawdb.NewLevelDBDatabase(*dbPath, 48, 48, "", false)

	if err != nil {
		return err
	}

	currentHead := rawdb.ReadHeadBlockHash(chainDb)

	if currentHead == (common.Hash{}) {
		return errors.New("we think head is genesis - an error with db most likely")
	}

	vmcfg := vm.Config{}

	engine := ethash.New(ethash.Config{
		CachesInMem:      ethconfig.Defaults.Ethash.CachesInMem,
		CachesOnDisk:     ethconfig.Defaults.Ethash.CachesOnDisk,
		CachesLockMmap:   ethconfig.Defaults.Ethash.CachesLockMmap,
		DatasetsInMem:    ethconfig.Defaults.Ethash.DatasetsInMem,
		DatasetsOnDisk:   ethconfig.Defaults.Ethash.DatasetsOnDisk,
		DatasetsLockMmap: ethconfig.Defaults.Ethash.DatasetsLockMmap,
	}, nil, false)

	cache := &core.CacheConfig{
		TrieCleanLimit: ethconfig.Defaults.TrieCleanCache,
		TrieDirtyLimit: ethconfig.Defaults.TrieDirtyCache,
		TrieTimeLimit:  ethconfig.Defaults.TrieTimeout,
		SnapshotLimit:  0,
	}

	chain, err := core.NewBlockChain(
		chainDb, cache, params.MainnetChainConfig, engine, vmcfg, nil, nil,
	)

	if err != nil {
		log.Fatal(err)
	}

	header := chain.CurrentHeader()
	blk := chain.GetBlockByNumber(uint64(blkNum))
	statedb, err := chain.StateAt(blk.Root())

	if err != nil {
		return err
	}

	blkCtx := core.NewEVMBlockContext(header, chain, nil)
	oneTimeEVM := vm.NewEVM(
		blkCtx, vm.TxContext{
			Origin:   sender,
			GasPrice: big.NewInt(int64(*gasPrice)),
		}, statedb, chain.Config(), *chain.GetVMConfig(),
	)

	result, _, err := oneTimeEVM.Call(
		vm.AccountRef(sender), toAddr, inputRunning,
		uint64(*gasLimit), common.Big0,
	)

	if err != nil {
		fmt.Println("error on evaling this message", err)
	} else {
		fmt.Println("output of call", hexutil.Encode(result))
	}
	return nil
}

func main() {
	flag.Parse()
	if err := program(); err != nil {
		log.Fatal(err)
	}
}
