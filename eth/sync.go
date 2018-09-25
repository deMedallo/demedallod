// Copyright 2015 The go-ethereum Authors
// This file is part of deMedallo.
//
// deMedallo is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// deMedallo is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with deMedallo. If not, see <http://www.gnu.org/licenses/>.

package eth

import (
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/demedallo/demedallod/common"
	"github.com/demedallo/demedallod/core/types"
	"github.com/demedallo/demedallod/eth/downloader"
	"github.com/demedallo/demedallod/logger"
	"github.com/demedallo/demedallod/logger/glog"
	"github.com/demedallo/demedallod/p2p/discover"
)

const (
	minDesiredPeerCount = 5 // Amount of peers desired to start syncing

	// This is the target size for the packs of transactions sent by txsyncLoop.
	// A pack can get larger than this if a single transactions exceeds this size.
	txsyncPackSize = 100 * 1024
)

type txsync struct {
	p   *peer
	txs []*types.Transaction
}

// syncTransactions starts sending all currently pending transactions to the given peer.
func (pm *ProtocolManager) syncTransactions(p *peer) {
	txs := pm.txpool.GetTransactions()
	if len(txs) == 0 {
		return
	}
	select {
	case pm.txsyncCh <- &txsync{p, txs}:
	case <-pm.quitSync:
	}
}

// txsyncLoop takes care of the initial transaction sync for each new
// connection. When a new peer appears, we relay all currently pending
// transactions. In order to minimise egress bandwidth usage, we send
// the transactions in small packs to one peer at a time.
func (pm *ProtocolManager) txsyncLoop() {
	var (
		pending = make(map[discover.NodeID]*txsync)
		sending = false               // whether a send is active
		pack    = new(txsync)         // the pack that is being sent
		done    = make(chan error, 1) // result of the send
	)

	// send starts a sending a pack of transactions from the sync.
	send := func(s *txsync) {
		// Fill pack with transactions up to the target size.
		size := common.StorageSize(0)
		pack.p = s.p
		pack.txs = pack.txs[:0]
		for i := 0; i < len(s.txs) && size < txsyncPackSize; i++ {
			pack.txs = append(pack.txs, s.txs[i])
			size += s.txs[i].Size()
		}
		// Remove the transactions that will be sent.
		s.txs = s.txs[:copy(s.txs, s.txs[len(pack.txs):])]
		if len(s.txs) == 0 {
			delete(pending, s.p.ID())
		}
		// Send the pack in the background.
		glog.V(logger.Detail).Infof("%v: sending %d transactions (%v)", s.p.Peer, len(pack.txs), size)
		sending = true
		go func() { done <- pack.p.SendTransactions(pack.txs) }()
	}

	// pick chooses the next pending sync.
	pick := func() *txsync {
		if len(pending) == 0 {
			return nil
		}
		n := rand.Intn(len(pending)) + 1
		for _, s := range pending {
			if n--; n == 0 {
				return s
			}
		}
		return nil
	}

	for {
		select {
		case s := <-pm.txsyncCh:
			pending[s.p.ID()] = s
			if !sending {
				send(s)
			}
		case err := <-done:
			sending = false
			// Stop tracking peers that cause send failures.
			if err != nil {
				glog.V(logger.Debug).Infof("%v: tx send failed: %v", pack.p.Peer, err)
				delete(pending, pack.p.ID())
			}
			// Schedule the next send.
			if s := pick(); s != nil {
				send(s)
			}
		case <-pm.quitSync:
			return
		}
	}
}

// syncer is responsible for periodically synchronising with the network, both
// downloading hashes and blocks as well as handling the announcement handler.
func (pm *ProtocolManager) syncer() {
	pm.fetcher.Start()
	defer pm.fetcher.Stop()
	defer pm.downloader.Terminate()

	sync := func() { pm.synchronise(pm.peers.BestPeer()) }
	for {
		batchTimer := time.AfterFunc(10*time.Second, sync)
		for {
			select {
			case <-pm.noMorePeers:
				batchTimer.Stop()
				return

			case <-pm.newPeerCh:
				if pm.peers.Len() < minDesiredPeerCount {
					continue
				}

				if batchTimer.Stop() {
					go sync()
				}

			case <-batchTimer.C:

			}
			// sync launched

			break
		}
	}
}

// synchronise tries to sync up our local block chain with a remote peer.
func (pm *ProtocolManager) synchronise(peer *peer) {
	// Short circuit if no peers are available
	if peer == nil {
		return
	}

	// Make sure the peer's TD is higher than our own
	currentBlock := pm.blockchain.CurrentBlock()
	td := pm.blockchain.GetTd(currentBlock.Hash())
	// Stored block's td should never be nil or non-positive.
	if td == nil || td.Sign() < 1 {
		glog.Fatalf("Found invalid TD=%v for current block in database. Exiting.\nCheck available disk space and restart to attempt database recovery.", td)
	}
	pHead, pTd := peer.Head()
	if pTd.Cmp(td) <= 0 {
		return
	}

	// Otherwise try to sync with the downloader
	mode := downloader.FullSync
	if atomic.LoadUint32(&pm.fastSync) == 1 {
		mode = downloader.FastSync
	}
	if !pm.downloader.Synchronise(peer.id, pHead, pTd, mode) {
		return
	}
	atomic.StoreUint32(&pm.synced, 1) // Mark initial sync done

	// If fast sync was enabled, and we synced up, disable it
	if atomic.LoadUint32(&pm.fastSync) == 1 {
		// Disable fast sync if we indeed have something in our chain
		if pm.blockchain.CurrentBlock().NumberU64() > 0 {
			glog.V(logger.Info).Infof("fast sync complete, auto disabling")
			atomic.StoreUint32(&pm.fastSync, 0)
		}
	}
}
