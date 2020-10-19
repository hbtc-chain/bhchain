package reactor

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/hbtc-chain/bhchain/services/p2p/pb"

	"github.com/stretchr/testify/assert"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/p2p"
	"google.golang.org/grpc"
)

const (
	basePort = 12330
)

func TestBroadcastBHMsg(t *testing.T) {
	n := 11
	reactors, msgChs := startBHNet(n)
	defer stopBHNet(log.TestingLogger(), reactors)

	msg := getTestMsg()
	reactors[0].localBHMsgFeed.Send(msg)

	assert.Equal(t, 0, verifyBHMsgReceived(t, msgChs[0], msg))
	for i := 1; i < n; i++ {
		assert.Equal(t, 1, verifyBHMsgReceived(t, msgChs[i], msg))
	}
}

func TestRelayMsgViaGrpc(t *testing.T) {
	n := 11
	reactors, msgChs := startBHNet(n)
	defer stopBHNet(log.TestingLogger(), reactors)

	startGrpcServer(t, reactors)

	msg := getTestMsg()
	reactors[0].localBHMsgFeed.Send(msg)

	for i := 0; i < n; i++ {
		assert.Equal(t, i, verifyBHMsgReceived(t, msgChs[i], msg))
	}
}

func startBHNet(n int) ([]*BHReactor, []chan *pb.BHMsg) {
	reactors := make([]*BHReactor, n)
	for i := 0; i < n; i++ {
		reactors[i] = NewBHReactor(log.NewFilter(log.TestingLogger(), log.AllowError()))
		reactors[i].SetLogger(log.TestingLogger().With("module", "bhmsg"))
	}
	// make connected switches and start all reactors
	p2p.MakeConnectedSwitches(cfg.TestConfig().P2P, n, func(i int, s *p2p.Switch) *p2p.Switch {
		s.AddReactor("BHReactor", reactors[i])
		s.SetLogger(log.TestingLogger().With("module", "p2p"))
		return s
	}, p2p.Connect2Switches)

	msgChs := make([]chan *pb.BHMsg, n)
	for i := 0; i < n; i++ {
		msgCh := make(chan *pb.BHMsg)
		reactors[i].incomingBHMsgFeed.Subscribe(msgCh)
		msgChs[i] = msgCh
	}

	return reactors, msgChs
}

func stopBHNet(logger log.Logger, reactors []*BHReactor) {
	logger.Info("stopBHNet", "n", len(reactors))
	for i, r := range reactors {
		logger.Info("stopBHNet: Stopping BHReactor", "i", i)
		r.Switch.Stop()
	}
	logger.Info("stopBHNet: DONE", "n", len(reactors))
}

func startGrpcServer(t *testing.T, reactors []*BHReactor) {
	servers := make([]*grpc.Server, len(reactors))
	for i, reactor := range reactors {
		i := i
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", basePort+i))
		assert.Nil(t, err)
		servers[i] = grpc.NewServer()
		pb.RegisterP2PServiceServer(servers[i], reactor)
		go func() {
			servers[i].Serve(lis)
		}()

		conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", basePort+i), grpc.WithInsecure())
		client := pb.NewP2PServiceClient(conn)

		stream, err := client.Communicate(context.Background())
		assert.Nil(t, err)
		go func() {
			for {
				msg, err := stream.Recv()
				if err == io.EOF {
					return
				}
				assert.Nil(t, err)
				time.Sleep(time.Duration(i*50) * time.Millisecond)
				assert.Nil(t, stream.Send(msg))
			}
		}()
	}

}

func getTestMsg() *pb.BHMsg {
	testMsg := &pb.BHMsg{
		Payload:   []byte("test_payload"),
		Type:      1,
		Signature: []byte("test_sig"),
	}

	return testMsg
}

func verifyBHMsgReceived(t *testing.T, ch chan *pb.BHMsg, expect *pb.BHMsg) int {
	var receivedCount int
out:
	for {
		select {
		case msg := <-ch:
			assert.EqualValues(t, expect.Signature, msg.Signature)
			assert.EqualValues(t, expect.Payload, msg.Payload)
			assert.EqualValues(t, expect.Type, msg.Type)
			receivedCount++
		case <-time.After(100 * time.Millisecond):
			break out
		}
	}
	return receivedCount
}
