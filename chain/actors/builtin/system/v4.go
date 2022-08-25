package system

import (
	"fmt"

	"github.com/ipfs/go-cid"

	system4 "github.com/filecoin-project/specs-actors/v4/actors/builtin/system"

	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/actors/adt"
)

var _ State = (*state4)(nil)

func load4(store adt.Store, root cid.Cid) (State, error) {
	out := state4{store: store}
	err := store.Get(store.Context(), root, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func make4(store adt.Store) (State, error) {
	out := state4{store: store}
	out.State = system4.State{}
	return &out, nil
}

type state4 struct {
	system4.State
	store adt.Store
}

func (s *state4) GetState() interface{} {
	return &s.State
}

func (s *state4) GetBuiltinActors() cid.Cid {

	return cid.Undef

}

func (s *state4) ActorKey() string {
	return actors.SystemKey
}

func (s *state4) ActorVersion() actors.Version {
	return actors.Version4
}

func (s *state4) Code() cid.Cid {
	code, ok := actors.GetActorCodeID(s.ActorVersion(), s.ActorKey())
	if !ok {
		panic(fmt.Errorf("didn't find actor %v code id for actor version %d", s.ActorKey(), s.ActorVersion()))
	}

	return code
}
