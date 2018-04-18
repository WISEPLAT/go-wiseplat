// Copyright 2016 The go-wiseplat Authors
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

// Package les implements the Light Wiseplat Subprotocol.
package les

import (
	"fmt"
	"sync"
	"time"

	"github.com/wiseplat/go-wiseplat/accounts"
	"github.com/wiseplat/go-wiseplat/common"
	"github.com/wiseplat/go-wiseplat/common/hexutil"
	"github.com/wiseplat/go-wiseplat/consensus"
	"github.com/wiseplat/go-wiseplat/core"
	"github.com/wiseplat/go-wiseplat/core/bloombits"
	"github.com/wiseplat/go-wiseplat/core/types"
	"github.com/wiseplat/go-wiseplat/wsh"
	"github.com/wiseplat/go-wiseplat/wsh/downloader"
	"github.com/wiseplat/go-wiseplat/wsh/filters"
	"github.com/wiseplat/go-wiseplat/wsh/gasprice"
	"github.com/wiseplat/go-wiseplat/wshdb"
	"github.com/wiseplat/go-wiseplat/event"
	"github.com/wiseplat/go-wiseplat/internal/wshapi"
	"github.com/wiseplat/go-wiseplat/light"
	"github.com/wiseplat/go-wiseplat/log"
	"github.com/wiseplat/go-wiseplat/node"
	"github.com/wiseplat/go-wiseplat/p2p"
	"github.com/wiseplat/go-wiseplat/p2p/discv5"
	"github.com/wiseplat/go-wiseplat/params"
	rpc "github.com/wiseplat/go-wiseplat/rpc"
)

type LightWiseplat struct {
	odr         *LesOdr
	relay       *LesTxRelay
	chainConfig *params.ChainConfig
	// Channel for shutting down the service
	shutdownChan chan bool
	// Handlers
	peers           *peerSet
	txPool          *light.TxPool
	blockchain      *light.LightChain
	protocolManager *ProtocolManager
	serverPool      *serverPool
	reqDist         *requestDistributor
	retriever       *retrieveManager
	// DB interfaces
	chainDb wshdb.Database // Block chain database

	bloomRequests                              chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer, chtIndexer, bloomTrieIndexer *core.ChainIndexer

	ApiBackend *LesApiBackend

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	networkId     uint64
	netRPCService *wshapi.PublicNetAPI

	wg sync.WaitGroup
}

func New(ctx *node.ServiceContext, config *wsh.Config) (*LightWiseplat, error) {
	chainDb, err := wsh.CreateDB(ctx, config, "lightchaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, isCompat := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !isCompat {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	peers := newPeerSet()
	quitSync := make(chan struct{})

	lwsh := &LightWiseplat{
		chainConfig:      chainConfig,
		chainDb:          chainDb,
		eventMux:         ctx.EventMux,
		peers:            peers,
		reqDist:          newRequestDistributor(peers, quitSync),
		accountManager:   ctx.AccountManager,
		engine:           wsh.CreateConsensusEngine(ctx, config, chainConfig, chainDb),
		shutdownChan:     make(chan bool),
		networkId:        config.NetworkId,
		bloomRequests:    make(chan chan *bloombits.Retrieval),
		bloomIndexer:     wsh.NewBloomIndexer(chainDb, light.BloomTrieFrequency),
		chtIndexer:       light.NewChtIndexer(chainDb, true),
		bloomTrieIndexer: light.NewBloomTrieIndexer(chainDb, true),
	}

	lwsh.relay = NewLesTxRelay(peers, lwsh.reqDist)
	lwsh.serverPool = newServerPool(chainDb, quitSync, &lwsh.wg)
	lwsh.retriever = newRetrieveManager(peers, lwsh.reqDist, lwsh.serverPool)
	lwsh.odr = NewLesOdr(chainDb, lwsh.chtIndexer, lwsh.bloomTrieIndexer, lwsh.bloomIndexer, lwsh.retriever)
	if lwsh.blockchain, err = light.NewLightChain(lwsh.odr, lwsh.chainConfig, lwsh.engine); err != nil {
		return nil, err
	}
	lwsh.bloomIndexer.Start(lwsh.blockchain)
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		lwsh.blockchain.SetHead(compat.RewindTo)
		core.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	lwsh.txPool = light.NewTxPool(lwsh.chainConfig, lwsh.blockchain, lwsh.relay)
	if lwsh.protocolManager, err = NewProtocolManager(lwsh.chainConfig, true, ClientProtocolVersions, config.NetworkId, lwsh.eventMux, lwsh.engine, lwsh.peers, lwsh.blockchain, nil, chainDb, lwsh.odr, lwsh.relay, quitSync, &lwsh.wg); err != nil {
		return nil, err
	}
	lwsh.ApiBackend = &LesApiBackend{lwsh, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	lwsh.ApiBackend.gpo = gasprice.NewOracle(lwsh.ApiBackend, gpoParams)
	return lwsh, nil
}

func lesTopic(genesisHash common.Hash, protocolVersion uint) discv5.Topic {
	var name string
	switch protocolVersion {
	case lpv1:
		name = "LES"
	case lpv2:
		name = "LES2"
	default:
		panic(nil)
	}
	return discv5.Topic(name + "@" + common.Bytes2Hex(genesisHash.Bytes()[0:8]))
}

type LightDummyAPI struct{}

// Wisebase is the address that mining rewards will be send to
func (s *LightDummyAPI) Wisebase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Coinbase is the address that mining rewards will be send to (alias for Wisebase)
func (s *LightDummyAPI) Coinbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Hashrate returns the POW hashrate
func (s *LightDummyAPI) Hashrate() hexutil.Uint {
	return 0
}

// Mining returns an indication if this node is currently mining.
func (s *LightDummyAPI) Mining() bool {
	return false
}

// APIs returns the collection of RPC services the wiseplat package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *LightWiseplat) APIs() []rpc.API {
	return append(wshapi.GetAPIs(s.ApiBackend), []rpc.API{
		{
			Namespace: "wsh",
			Version:   "1.0",
			Service:   &LightDummyAPI{},
			Public:    true,
		}, {
			Namespace: "wsh",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "wsh",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, true),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *LightWiseplat) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *LightWiseplat) BlockChain() *light.LightChain      { return s.blockchain }
func (s *LightWiseplat) TxPool() *light.TxPool              { return s.txPool }
func (s *LightWiseplat) Engine() consensus.Engine           { return s.engine }
func (s *LightWiseplat) LesVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *LightWiseplat) Downloader() *downloader.Downloader { return s.protocolManager.downloader }
func (s *LightWiseplat) EventMux() *event.TypeMux           { return s.eventMux }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *LightWiseplat) Protocols() []p2p.Protocol {
	return s.protocolManager.SubProtocols
}

// Start implements node.Service, starting all internal goroutines needed by the
// Wiseplat protocol implementation.
func (s *LightWiseplat) Start(srvr *p2p.Server) error {
	s.startBloomHandlers()
	log.Warn("Light client mode is an experimental feature")
	s.netRPCService = wshapi.NewPublicNetAPI(srvr, s.networkId)
	// search the topic belonging to the oldest supported protocol because
	// servers always advertise all supported protocols
	protocolVersion := ClientProtocolVersions[len(ClientProtocolVersions)-1]
	s.serverPool.start(srvr, lesTopic(s.blockchain.Genesis().Hash(), protocolVersion))
	s.protocolManager.Start()
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Wiseplat protocol.
func (s *LightWiseplat) Stop() error {
	s.odr.Stop()
	if s.bloomIndexer != nil {
		s.bloomIndexer.Close()
	}
	if s.chtIndexer != nil {
		s.chtIndexer.Close()
	}
	if s.bloomTrieIndexer != nil {
		s.bloomTrieIndexer.Close()
	}
	s.blockchain.Stop()
	s.protocolManager.Stop()
	s.txPool.Stop()

	s.eventMux.Stop()

	time.Sleep(time.Millisecond * 200)
	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}
