// Copyright 2015 The go-wiseplat Authors
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

package wsh

import (
	"context"
	"math/big"

	"github.com/wiseplat/go-wiseplat/accounts"
	"github.com/wiseplat/go-wiseplat/common"
	"github.com/wiseplat/go-wiseplat/common/math"
	"github.com/wiseplat/go-wiseplat/core"
	"github.com/wiseplat/go-wiseplat/core/bloombits"
	"github.com/wiseplat/go-wiseplat/core/state"
	"github.com/wiseplat/go-wiseplat/core/types"
	"github.com/wiseplat/go-wiseplat/core/vm"
	"github.com/wiseplat/go-wiseplat/wsh/downloader"
	"github.com/wiseplat/go-wiseplat/wsh/gasprice"
	"github.com/wiseplat/go-wiseplat/wshdb"
	"github.com/wiseplat/go-wiseplat/event"
	"github.com/wiseplat/go-wiseplat/params"
	"github.com/wiseplat/go-wiseplat/rpc"
)

// WshApiBackend implements wshapi.Backend for full nodes
type WshApiBackend struct {
	wsh *Wiseplat
	gpo *gasprice.Oracle
}

func (b *WshApiBackend) ChainConfig() *params.ChainConfig {
	return b.wsh.chainConfig
}

func (b *WshApiBackend) CurrentBlock() *types.Block {
	return b.wsh.blockchain.CurrentBlock()
}

func (b *WshApiBackend) SetHead(number uint64) {
	b.wsh.protocolManager.downloader.Cancel()
	b.wsh.blockchain.SetHead(number)
}

func (b *WshApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.wsh.miner.PendingBlock()
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.wsh.blockchain.CurrentBlock().Header(), nil
	}
	return b.wsh.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *WshApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.wsh.miner.PendingBlock()
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.wsh.blockchain.CurrentBlock(), nil
	}
	return b.wsh.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *WshApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.wsh.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.wsh.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *WshApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	return b.wsh.blockchain.GetBlockByHash(blockHash), nil
}

func (b *WshApiBackend) GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error) {
	return core.GetBlockReceipts(b.wsh.chainDb, blockHash, core.GetBlockNumber(b.wsh.chainDb, blockHash)), nil
}

func (b *WshApiBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.wsh.blockchain.GetTdByHash(blockHash)
}

func (b *WshApiBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.wsh.BlockChain(), nil)
	return vm.NewEVM(context, state, b.wsh.chainConfig, vmCfg), vmError, nil
}

func (b *WshApiBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.wsh.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *WshApiBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.wsh.BlockChain().SubscribeChainEvent(ch)
}

func (b *WshApiBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.wsh.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *WshApiBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.wsh.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *WshApiBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.wsh.BlockChain().SubscribeLogsEvent(ch)
}

func (b *WshApiBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.wsh.txPool.AddLocal(signedTx)
}

func (b *WshApiBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.wsh.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *WshApiBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.wsh.txPool.Get(hash)
}

func (b *WshApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.wsh.txPool.State().GetNonce(addr), nil
}

func (b *WshApiBackend) Stats() (pending int, queued int) {
	return b.wsh.txPool.Stats()
}

func (b *WshApiBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.wsh.TxPool().Content()
}

func (b *WshApiBackend) SubscribeTxPreEvent(ch chan<- core.TxPreEvent) event.Subscription {
	return b.wsh.TxPool().SubscribeTxPreEvent(ch)
}

func (b *WshApiBackend) Downloader() *downloader.Downloader {
	return b.wsh.Downloader()
}

func (b *WshApiBackend) ProtocolVersion() int {
	return b.wsh.WshVersion()
}

func (b *WshApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *WshApiBackend) ChainDb() wshdb.Database {
	return b.wsh.ChainDb()
}

func (b *WshApiBackend) EventMux() *event.TypeMux {
	return b.wsh.EventMux()
}

func (b *WshApiBackend) AccountManager() *accounts.Manager {
	return b.wsh.AccountManager()
}

func (b *WshApiBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.wsh.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *WshApiBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.wsh.bloomRequests)
	}
}
