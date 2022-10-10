package vm

import (
	"context"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/crypto"
	"os"

	cid "github.com/ipfs/go-cid"

	"github.com/filecoin-project/go-state-types/network"

	"github.com/filecoin-project/lotus/chain/types"
)

// stat counters
var (
	StatSends   uint64
	StatApplied uint64
)

type Interface interface {
	// Applies the given message onto the VM's current state, returning the result of the execution
	ApplyMessage(ctx context.Context, cmsg types.ChainMsg) (*ApplyRet, error)
	// Same as above but for system messages (the Cron invocation and block reward payments).
	// Must NEVER fail.
	ApplyImplicitMessage(ctx context.Context, msg *types.Message) (*ApplyRet, error)
	// Flush all buffered objects into the state store provided to the VM at construction.
	Flush(ctx context.Context) (cid.Cid, error)
}

// WARNING: You will not affect your node's execution by misusing this feature, but you will confuse yourself thoroughly!
// An envvar that allows the user to specify debug actors bundles to be used by the FVM
// alongside regular execution. This is basically only to be used to print out specific logging information.
// Message failures, unexpected terminations,gas costs, etc. should all be ignored.
var useFvmDebug = os.Getenv("LOTUS_FVM_DEVELOPER_DEBUG") == "1"

func NewVM(ctx context.Context, opts *VMOpts) (Interface, error) {
	// if this_is_for_adlrocha_to_run {
	opts.Rand = veryFakeRand{}
	// }
	if opts.NetworkVersion >= network.Version16 {
		if useFvmDebug {
			return NewDualExecutionFVM(ctx, opts)
		}
		return NewFVM(ctx, opts)
	}

	return NewLegacyVM(ctx, opts)
}

type veryFakeRand struct{}

func (v veryFakeRand) GetChainRandomness(ctx context.Context, pers crypto.DomainSeparationTag, round abi.ChainEpoch, entropy []byte) ([]byte, error) {
	return []byte("sorandomlol"), nil
}

func (v veryFakeRand) GetBeaconRandomness(ctx context.Context, pers crypto.DomainSeparationTag, round abi.ChainEpoch, entropy []byte) ([]byte, error) {
	return []byte("sorandomlol"), nil

}
