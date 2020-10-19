/*
 * *******************************************************************
 * @项目名称: reactor
 * @文件名称: reactor.go
 * @Date: 2019/04/19
 * @Author: kai.wen
 * @Copyright（C）: 2019 BlueHelix Inc.   All rights reserved.
 * 注意：本内容仅限于内部传阅，禁止外泄以及用于其他的商业目的.
 * *******************************************************************
 */

package reactor

import (
	"fmt"
	"io"
	"sync"

	"github.com/hbtc-chain/bhchain/services/p2p/event"
	"github.com/hbtc-chain/bhchain/services/p2p/pb"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/p2p"
)

const (
	bhmID = "m"

	// BHMsgChannel id
	BHMsgChannel = byte(0x70)

	maxMsgSize = 1048576 // 1MB; NOTE/TODO: keep in sync with types.PartSet sizes.

	maxMarkMapSize = 100000
)

var _ p2p.Reactor = (*BHReactor)(nil)
var _ pb.P2PServiceServer = (*BHReactor)(nil)

// BHReactor defines a reactor for the key manager service.
type BHReactor struct {
	p2p.BaseReactor // BaseService + p2p.Switch

	localBHMsgFeed    *event.Feed // message from grpc stream
	incomingBHMsgFeed *event.Feed // message from p2p networ

	peerMap map[p2p.ID]*peerWithMark

	mtx sync.RWMutex
}

type peerWithMark struct {
	p2p.Peer
	markedMap map[Hash]bool
	mtx       sync.RWMutex
}

func (p *peerWithMark) markBHMsg(hash Hash) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	if len(p.markedMap) >= maxMarkMapSize {
		p.markedMap = make(map[Hash]bool)
	}
	p.markedMap[hash] = true
}

func (p *peerWithMark) marked(hash Hash) bool {
	p.mtx.RLock()
	defer p.mtx.RUnlock()

	return p.markedMap[hash]
}

// NewBHReactor returns a new BHReactor.
func NewBHReactor(logger log.Logger) *BHReactor {

	bhR := &BHReactor{
		localBHMsgFeed:    &event.Feed{},
		incomingBHMsgFeed: &event.Feed{},
		peerMap:           make(map[p2p.ID]*peerWithMark),
	}
	bhR.BaseReactor = *p2p.NewBaseReactor("BHMsgReactor", bhR)

	return bhR
}

// OnStart implements BaseService by subscribing to events, which later will be
// broadcasted to other peers and starting state if we're not in fast sync.
func (bhR *BHReactor) OnStart() error {
	return nil
}

// OnStop implements BaseService by unsubscribing from events and stopping
// state.
func (bhR *BHReactor) OnStop() {
	// TODO(kai.wen): stop bhm
}

// GetChannels implements Reactor
func (bhR *BHReactor) GetChannels() []*p2p.ChannelDescriptor {
	return []*p2p.ChannelDescriptor{
		{
			ID:                  BHMsgChannel,
			Priority:            5,
			SendQueueCapacity:   100,
			RecvMessageCapacity: maxMsgSize,
		},
	}
}

// AddPeer implements Reactor
func (bhR *BHReactor) AddPeer(peer p2p.Peer) {
	if !bhR.IsRunning() {
		return
	}

	p := &peerWithMark{
		markedMap: make(map[Hash]bool),
		Peer:      peer,
	}

	bhR.mtx.Lock()
	bhR.peerMap[peer.ID()] = p
	bhR.mtx.Unlock()

	go bhR.gossipBHMsgRoutine(p)
}

func (bhR *BHReactor) gossipBHMsgRoutine(peer *peerWithMark) {
	logger := bhR.Logger.With("peer", peer)
	ch := make(chan *pb.BHMsg, 1)
	sub := bhR.localBHMsgFeed.Subscribe(ch)
	defer sub.Unsubscribe()
	for {
		// Manage disconnects from self or peer.
		if !peer.IsRunning() || !bhR.IsRunning() {
			logger.Info("Stopping gossipBHMsgRoutine for peer")
			return
		}

		select {
		case msg := <-ch:
			hash := toHash(msg.Payload)
			if !peer.marked(hash) {
				peer.Send(BHMsgChannel, cdc.MustMarshalBinaryBare(msg))
				peer.markBHMsg(hash)
			}
		}
	}
}

// RemovePeer implements Reactor
func (bhR *BHReactor) RemovePeer(peer p2p.Peer, reason interface{}) {
	if !bhR.IsRunning() {
		return
	}

	bhR.mtx.Lock()
	defer bhR.mtx.Unlock()
	delete(bhR.peerMap, peer.ID())
}

func decodeMsg(bz []byte) (*pb.BHMsg, error) {
	if len(bz) > maxMsgSize {
		return nil, fmt.Errorf("Msg exceeds max size (%d > %d)", len(bz), maxMsgSize)
	}
	var msg pb.BHMsg
	err := cdc.UnmarshalBinaryBare(bz, &msg)
	return &msg, err
}

// Receive implements Reactor
// NOTE: We process these messages even when we're fast_syncing.
// Messages affect either a peer state or the consensus state.
// Peer state updates can happen in parallel, but processing of
// proposals, block parts, and votes are ordered by the receiveRoutine
// NOTE: blocks on consensus state for proposals, block parts, and votes
func (bhR *BHReactor) Receive(chID byte, src p2p.Peer, msgBytes []byte) {
	if !bhR.IsRunning() {
		return
	}

	msg, err := decodeMsg(msgBytes)
	if err != nil {
		bhR.Switch.StopPeerForError(src, err)
		return
	}

	bhR.mtx.RLock()
	p, ok := bhR.peerMap[src.ID()]
	bhR.mtx.RUnlock()
	if !ok {
		bhR.Switch.StopPeerForError(src, err)
		return
	}

	p.markBHMsg(toHash(msg.Payload))

	bhR.incomingBHMsgFeed.Send(msg)
}

func (bhR *BHReactor) Communicate(stream pb.P2PService_CommunicateServer) error {
	errCh := make(chan error)
	outCh := make(chan *pb.BHMsg)
	outSub := bhR.incomingBHMsgFeed.Subscribe(outCh)
	defer outSub.Unsubscribe()
	go func() {
		for {
			in, err := stream.Recv()
			if err != nil {
				errCh <- err
				return
			}
			bhR.localBHMsgFeed.Send(in)
		}
	}()
	for {
		select {
		case err := <-errCh:
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
		case out := <-outCh:
			err := stream.Send(out)
			if err != nil {
				bhR.Logger.Error("Send message to grpc stream error", "err", err)
			}
		case err := <-outSub.Err():
			return err
		}
	}
}