package itests

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/filecoin-project/lotus/chain/messagesigner"
	"github.com/filecoin-project/lotus/itests/kit"
	consensus "github.com/filecoin-project/lotus/lib/consensus/raft"
	"github.com/filecoin-project/lotus/node"
	"github.com/filecoin-project/lotus/node/modules"
	gorpc "github.com/libp2p/go-libp2p-gorpc"
	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func generatePrivKey() (*kit.Libp2p, error) {
	privkey, _, err := libp2pcrypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, err
	}

	peerId, err := peer.IDFromPrivateKey(privkey)
	if err != nil {
		return nil, err
	}

	return &kit.Libp2p{peerId, privkey}, nil
}

// TestMultisig does a basic test to exercise the multisig CLI commands
func TestRaftMessageSigner(t *testing.T) {

	kit.QuietMiningLogs()
	ctx := context.Background()

	blockTime := 5 * time.Millisecond

	var (
		node0 kit.TestFullNode
		node1 kit.TestFullNode
		node2 kit.TestFullNode
		miner kit.TestMiner
	)

	pkey0, _ := generatePrivKey()
	pkey1, _ := generatePrivKey()
	pkey2, _ := generatePrivKey()

	initPeerSet := []peer.ID{pkey0.PeerID, pkey1.PeerID, pkey2.PeerID}

	raftOps := kit.ConstructorOpts(
		node.Override(new(*gorpc.Client), modules.NewRPCClient),
		node.Override(new(*consensus.Config), func() *consensus.Config {
			cfg := consensus.NewDefaultConfig()
			cfg.InitPeerset = initPeerSet
			return cfg
		}),
		node.Override(new(*consensus.Consensus), consensus.NewConsensusWithRPCClient(false)),
		node.Override(new(*messagesigner.MessageSignerConsensus), messagesigner.NewMessageSignerConsensus),
		node.Override(new(messagesigner.MsgSigner), func(ms *messagesigner.MessageSignerConsensus) *messagesigner.MessageSignerConsensus { return ms }),
		node.Override(new(*modules.RPCHandler), modules.NewRPCHandler),
		node.Override(node.RPCServer, modules.NewRPCServer),
	)
	//raftOps := kit.ConstructorOpts()

	ens := kit.NewEnsemble(t).FullNode(&node0, raftOps).FullNode(&node1, raftOps).FullNode(&node2, raftOps)
	node0.AssignPrivKey(pkey0)
	node1.AssignPrivKey(pkey1)
	node2.AssignPrivKey(pkey2)

	ens.MinerEnroll(&miner, &node0, kit.WithAllSubsystems())

	ens.Start()

	_, err := node0.WalletImport(ctx, &miner.OwnerKey.KeyInfo)
	require.NoError(t, err)

	_, err = node1.WalletImport(ctx, &miner.OwnerKey.KeyInfo)
	require.NoError(t, err)

	_, err = node2.WalletImport(ctx, &miner.OwnerKey.KeyInfo)
	require.NoError(t, err)

	ens.InterconnectAll()

	ens.AddInactiveMiner(&miner)
	ens.Start()

	//wallets, err := node0.WalletList(ctx)
	//require.NoError(t, err)
	//
	//walletKey, err := node0.WalletExport(ctx, wallets[0])
	//require.NoError(t, err)
	//
	//_, err = node1.WalletImport(ctx, walletKey)
	//require.NoError(t, err)
	//
	//_, err = node2.WalletImport(ctx, walletKey)
	//require.NoError(t, err)

	//client, _, ens := kit.EnsembleMinimal(t, kit.MockProofs(), kit.ThroughRPC())
	ens.InterconnectAll().BeginMining(blockTime)

	fmt.Println(node0.WalletList(context.Background()))
	fmt.Println(node1.WalletList(context.Background()))
	fmt.Println(node2.WalletList(context.Background()))

}
