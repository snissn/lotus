package messagesigner

import (
	"context"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	consensus "github.com/filecoin-project/lotus/lib/consensus/raft"
	"github.com/filecoin-project/lotus/node/modules/dtypes"
	"github.com/google/uuid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	"golang.org/x/xerrors"
	"reflect"
)

type MessageSignerConsensus struct {
	//msgSigner MessageSigner
	MsgSigner
	consensus *consensus.Consensus
}

//var _ full.MsgSigner = &MessageSignerConsensus{}

func NewMessageSignerConsensus(
	wallet api.Wallet,
	mpool MpoolNonceAPI,
	ds dtypes.MetadataDS,
	consensus *consensus.Consensus) *MessageSignerConsensus {

	ds = namespace.Wrap(ds, datastore.NewKey("/message-signer-consensus/"))
	return &MessageSignerConsensus{
		MsgSigner: &MessageSigner{
			wallet: wallet,
			mpool:  mpool,
			ds:     ds,
		},
		consensus: consensus,
	}
}

func (ms *MessageSignerConsensus) IsLeader(ctx context.Context) bool {
	return ms.consensus.IsLeader(ctx)
}

func (ms *MessageSignerConsensus) RedirectToLeader(ctx context.Context, method string, arg interface{}, ret interface{}) (bool, error) {
	ok, err := ms.consensus.RedirectToLeader(method, arg, ret.(*types.SignedMessage))
	if err != nil {
		return ok, err
	}
	return ok, nil
}

func (ms *MessageSignerConsensus) SignMessage(
	ctx context.Context,
	msg *types.Message,
	spec *api.MessageSendSpec,
	cb func(*types.SignedMessage) error) (*types.SignedMessage, error) {

	signedMsg, err := ms.MsgSigner.SignMessage(ctx, msg, spec, cb)
	if err != nil {
		return nil, err
	}
	//u := uuid.New()
	//if spec != nil {
	//	u = spec.MsgUuid
	//}

	//leader, err := ms.consensus.Leader(ctx)
	////curr := ms.consensus.IsLeader(ctx)
	//log.Infof("Consensus leader: ", leader, "current node is leader: ", ms.consensus.IsLeader(ctx))

	op := &consensus.ConsensusOp{signedMsg.Message.Nonce, spec.MsgUuid, signedMsg.Message.From, signedMsg}
	err = ms.consensus.Commit(ctx, op)
	if err != nil {
		return nil, err
	}

	return signedMsg, nil
}

func (ms *MessageSignerConsensus) GetSignedMessage(ctx context.Context, uuid uuid.UUID) (*types.SignedMessage, error) {
	state, err := ms.consensus.State(ctx)
	if err != nil {
		return nil, err
	}
	//ev := reflect.ValueOf(state)
	//et := ev.Type()
	log.Infof("!!!!!!!!!!!!!!!!!!!!!!!STate type: %v", reflect.TypeOf(state))

	cstate := state.(*consensus.RaftState)
	msg, ok := cstate.MsgUuids[uuid]
	if !ok {
		return nil, xerrors.Errorf("Msg with Uuid %s not available", uuid)
	}
	return msg, nil
}

//func (ms *MessageSignerConsensus) StoreSignedMessage(ctx context.Context, uuid uuid.UUID, message *types.SignedMessage) error {
//
//	ms.consensus
//	return ms.MsgSigner.StoreSignedMessage(ctx, uuid, message)
//}
//
//func (ms *MessageSignerConsensus) NextNonce(ctx context.Context, addr address.Address) (uint64, error) {
//	return ms.msgSigner.NextNonce(ctx, addr)
//}
//
//func (ms *MessageSignerConsensus) SaveNonce(ctx context.Context, addr address.Address, nonce uint64) error {
//	return ms.msgSigner.SaveNonce(ctx, addr, nonce)
//}
//
//func (ms *MessageSignerConsensus) DstoreKey(addr address.Address) datastore.Key {
//	return ms.msgSigner.DstoreKey(addr)
//}
