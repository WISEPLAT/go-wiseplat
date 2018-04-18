// Copyright 2014 The go-wiseplat Authors
// This file is part of the go-wiseplat library.
//
// The go-wiseplat library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-wiseplat library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-wiseplat library. If not, see <http://www.gnu.org/licenses/>.

// Package wsh implements the Wiseplat protocol.
package wsh

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/wiseplat/go-wiseplat/accounts"
	"github.com/wiseplat/go-wiseplat/common"
	"github.com/wiseplat/go-wiseplat/common/hexutil"
	"github.com/wiseplat/go-wiseplat/consensus"
	"github.com/wiseplat/go-wiseplat/consensus/clique"
	"github.com/wiseplat/go-wiseplat/consensus/wshash"
	"github.com/wiseplat/go-wiseplat/core"
	"github.com/wiseplat/go-wiseplat/core/bloombits"
	"github.com/wiseplat/go-wiseplat/core/types"
	"github.com/wiseplat/go-wiseplat/core/vm"
	"github.com/wiseplat/go-wiseplat/wsh/downloader"
	"github.com/wiseplat/go-wiseplat/wsh/filters"
	"github.com/wiseplat/go-wiseplat/wsh/gasprice"
	"github.com/wiseplat/go-wiseplat/wshdb"
	"github.com/wiseplat/go-wiseplat/event"
	"github.com/wiseplat/go-wiseplat/internal/wshapi"
	"github.com/wiseplat/go-wiseplat/log"
	"github.com/wiseplat/go-wiseplat/miner"
	"github.com/wiseplat/go-wiseplat/node"
	"github.com/wiseplat/go-wiseplat/p2p"
	"github.com/wiseplat/go-wiseplat/params"
	"github.com/wiseplat/go-wiseplat/rlp"
	"github.com/wiseplat/go-wiseplat/rpc"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *core.ChainIndexer)
}

// Wiseplat implements the Wiseplat full node service.
type Wiseplat struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan  chan bool    // Channel for shutting down the wiseplat
	stopDbUpgrade func() error // stop chain db sequential key upgrade

	// Handlers
	txPool          *core.TxPool
	blockchain      *core.BlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer

	// DB interfaces
	chainDb wshdb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	ApiBackend *WshApiBackend

	miner     *miner.Miner
	gasPrice  *big.Int
	wisebase common.Address

	networkId     uint64
	netRPCService *wshapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and wisebase)
}

func (s *Wiseplat) AddLesServer(ls LesServer) {
	s.lesServer = ls
	ls.SetBloomBitsIndexer(s.bloomIndexer)
}

// New creates a new Wiseplat object (including the
// initialisation of the common Wiseplat object)
func New(ctx *node.ServiceContext, config *Config) (*Wiseplat, error) {
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run wsh.Wiseplat in light sync mode, use les.LightWiseplat")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	stopDbUpgrade := upgradeDeduplicateData(chainDb)
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	wsh := &Wiseplat{
		config:         config,
		chainDb:        chainDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		engine:         CreateConsensusEngine(ctx, config, chainConfig, chainDb),
		shutdownChan:   make(chan bool),
		stopDbUpgrade:  stopDbUpgrade,
		networkId:      config.NetworkId,
		gasPrice:       config.GasPrice,
		wisebase:      config.Wisebase,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   NewBloomIndexer(chainDb, params.BloomBitsBlocks),
	}

	log.Info("Initialising Wiseplat protocol", "versions", ProtocolVersions, "network", config.NetworkId)

	if !config.SkipBcVersionCheck {
		bcVersion := core.GetBlockChainVersion(chainDb)
		if bcVersion != core.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run gwsh upgradedb.\n", bcVersion, core.BlockChainVersion)
		}
		core.WriteBlockChainVersion(chainDb, core.BlockChainVersion)
	}

	vmConfig := vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
	wsh.blockchain, err = core.NewBlockChain(chainDb, wsh.chainConfig, wsh.engine, vmConfig)
	if err != nil {
		return nil, err
	}
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		wsh.blockchain.SetHead(compat.RewindTo)
		core.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	wsh.bloomIndexer.Start(wsh.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	wsh.txPool = core.NewTxPool(config.TxPool, wsh.chainConfig, wsh.blockchain)

	if wsh.protocolManager, err = NewProtocolManager(wsh.chainConfig, config.SyncMode, config.NetworkId, wsh.eventMux, wsh.txPool, wsh.engine, wsh.blockchain, chainDb); err != nil {
		return nil, err
	}
	wsh.miner = miner.New(wsh, wsh.chainConfig, wsh.EventMux(), wsh.engine)
	wsh.miner.SetExtra(makeExtraData(config.ExtraData))

	wsh.ApiBackend = &WshApiBackend{wsh, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	wsh.ApiBackend.gpo = gasprice.NewOracle(wsh.ApiBackend, gpoParams)

	return wsh, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"gwsh",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (wshdb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if db, ok := db.(*wshdb.LDBDatabase); ok {
		db.Meter("wsh/db/chaindata/")
	}
	return db, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an Wiseplat service
func CreateConsensusEngine(ctx *node.ServiceContext, config *Config, chainConfig *params.ChainConfig, db wshdb.Database) consensus.Engine {
	// If proof-of-authority is requested, set it up
	if chainConfig.Clique != nil {
		return clique.New(chainConfig.Clique, db)
	}
	// Otherwise assume proof-of-work
	switch {
	case config.PowFake:
		log.Warn("Wshash used in fake mode")
		return wshash.NewFaker()
	case config.PowTest:
		log.Warn("Wshash used in test mode")
		return wshash.NewTester()
	case config.PowShared:
		log.Warn("Wshash used in shared mode")
		return wshash.NewShared()
	default:
		engine := wshash.New(ctx.ResolvePath(config.WshashCacheDir), config.WshashCachesInMem, config.WshashCachesOnDisk,
			config.WshashDatasetDir, config.WshashDatasetsInMem, config.WshashDatasetsOnDisk)
		engine.SetThreads(-1) // Disable CPU mining
		return engine
	}
}

// APIs returns the collection of RPC services the wiseplat package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Wiseplat) APIs() []rpc.API {
	apis := wshapi.GetAPIs(s.ApiBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)

	// Append all the local APIs and return
	return append(apis, []rpc.API{
		{
			Namespace: "wsh",
			Version:   "1.0",
			Service:   NewPublicWiseplatAPI(s),
			Public:    true,
		}, {
			Namespace: "wsh",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(s),
			Public:    true,
		}, {
			Namespace: "wsh",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(s),
			Public:    false,
		}, {
			Namespace: "wsh",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, false),
			Public:    true,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(s),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(s),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(s.chainConfig, s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *Wiseplat) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Wiseplat) Wisebase() (eb common.Address, err error) {
	s.lock.RLock()
	wisebase := s.wisebase
	s.lock.RUnlock()

	if wisebase != (common.Address{}) {
		return wisebase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			return accounts[0].Address, nil
		}
	}
	return common.Address{}, fmt.Errorf("wisebase address must be explicitly specified")
}

// set in js console via admin interface or wrapper from cli flags
func (self *Wiseplat) SetWisebase(wisebase common.Address) {
	self.lock.Lock()
	self.wisebase = wisebase
	self.lock.Unlock()

	self.miner.SetWisebase(wisebase)
}

func (s *Wiseplat) StartMining(local bool) error {
	eb, err := s.Wisebase()
	if err != nil {
		log.Error("Cannot start mining without wisebase", "err", err)
		return fmt.Errorf("wisebase missing: %v", err)
	}
	if clique, ok := s.engine.(*clique.Clique); ok {
		wallet, err := s.accountManager.Find(accounts.Account{Address: eb})
		if wallet == nil || err != nil {
			log.Error("Wisebase account unavailable locally", "err", err)
			return fmt.Errorf("signer missing: %v", err)
		}
		clique.Authorize(eb, wallet.SignHash)
	}
	if local {
		// If local (CPU) mining is started, we can disable the transaction rejection
		// mechanism introduced to speed sync times. CPU mining on mainnet is ludicrous
		// so noone will ever hit this path, whereas marking sync done on CPU mining
		// will ensure that private networks work in single miner mode too.
		atomic.StoreUint32(&s.protocolManager.acceptTxs, 1)
	}
	go s.miner.Start(eb)
	return nil
}

func (s *Wiseplat) StopMining()         { s.miner.Stop() }
func (s *Wiseplat) IsMining() bool      { return s.miner.Mining() }
func (s *Wiseplat) Miner() *miner.Miner { return s.miner }

func (s *Wiseplat) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *Wiseplat) BlockChain() *core.BlockChain       { return s.blockchain }
func (s *Wiseplat) TxPool() *core.TxPool               { return s.txPool }
func (s *Wiseplat) EventMux() *event.TypeMux           { return s.eventMux }
func (s *Wiseplat) Engine() consensus.Engine           { return s.engine }
func (s *Wiseplat) ChainDb() wshdb.Database            { return s.chainDb }
func (s *Wiseplat) IsListening() bool                  { return true } // Always listening
func (s *Wiseplat) WshVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *Wiseplat) NetVersion() uint64                 { return s.networkId }
func (s *Wiseplat) Downloader() *downloader.Downloader { return s.protocolManager.downloader }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *Wiseplat) Protocols() []p2p.Protocol {
	if s.lesServer == nil {
		return s.protocolManager.SubProtocols
	}
	return append(s.protocolManager.SubProtocols, s.lesServer.Protocols()...)
}

// Start implements node.Service, starting all internal goroutines needed by the
// Wiseplat protocol implementation.
func (s *Wiseplat) Start(srvr *p2p.Server) error {
	// Start the bloom bits servicing goroutines
	s.startBloomHandlers()

	// Start the RPC service
	s.netRPCService = wshapi.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	if s.config.LightServ > 0 {
		maxPeers -= s.config.LightPeers
		if maxPeers < srvr.MaxPeers/2 {
			maxPeers = srvr.MaxPeers / 2
		}
	}
	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	if s.lesServer != nil {
		s.lesServer.Start(srvr)
	}
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Wiseplat protocol.
func (s *Wiseplat) Stop() error {
	if s.stopDbUpgrade != nil {
		s.stopDbUpgrade()
	}
	s.bloomIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	if s.lesServer != nil {
		s.lesServer.Stop()
	}
	s.txPool.Stop()
	s.miner.Stop()
	s.eventMux.Stop()

	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}
