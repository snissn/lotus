package main

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/filecoin-project/go-state-types/abi"

	"github.com/ipfs/go-cid"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	actorstypes "github.com/filecoin-project/go-state-types/actors"
	"github.com/filecoin-project/go-state-types/builtin"
	v8 "github.com/filecoin-project/go-state-types/builtin/v8"
	v9 "github.com/filecoin-project/go-state-types/builtin/v9"

	"github.com/filecoin-project/lotus/blockstore"
	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/consensus/filcns"
	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/chain/types"
	lcli "github.com/filecoin-project/lotus/cli"
	"github.com/filecoin-project/lotus/node/repo"
)

var invariantsCmd = &cli.Command{
	Name:        "check-invariants",
	Description: "Check state invariants",
	ArgsUsage:   "[stateRoot height]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "repo",
			Value: "~/.lotus",
		},
	},
	Action: func(cctx *cli.Context) error {
		ctx := context.TODO()

		if cctx.NArg() != 2 {
			return lcli.IncorrectNumArgs(cctx)
		}

		rootCid, err := cid.Decode(cctx.Args().First())
		if err != nil {
			return fmt.Errorf("failed to parse input: %w", err)
		}

		height, err := strconv.ParseUint(cctx.Args().Get(1), 10, 64)

		fsrepo, err := repo.NewFS(cctx.String("repo"))
		if err != nil {
			return err
		}

		lkrepo, err := fsrepo.Lock(repo.FullNode)
		if err != nil {
			return err
		}

		defer lkrepo.Close() //nolint:errcheck

		bs, err := lkrepo.Blockstore(ctx, repo.UniversalBlockstore)
		if err != nil {
			return fmt.Errorf("failed to open blockstore: %w", err)
		}

		defer func() {
			if c, ok := bs.(io.Closer); ok {
				if err := c.Close(); err != nil {
					log.Warnf("failed to close blockstore: %s", err)
				}
			}
		}()

		mds, err := lkrepo.Datastore(context.Background(), "/metadata")
		if err != nil {
			return err
		}

		cs := store.NewChainStore(bs, bs, mds, filcns.Weight, nil)
		defer cs.Close() //nolint:errcheck

		av := actorstypes.Version9
		fmt.Println("Actors Version ", av)

		actorCodeCids, err := actors.GetActorCodeIDs(av)
		if err != nil {
			return err
		}

		actorStore := store.ActorStore(ctx, blockstore.NewTieredBstore(bs, blockstore.NewMemorySync()))

		// Load the state root.
		var stateRoot types.StateRoot
		if err := actorStore.Get(ctx, rootCid, &stateRoot); err != nil {
			return xerrors.Errorf("failed to decode state root: %w", err)
		}

		actorTree, err := builtin.LoadTree(actorStore, stateRoot.Actors)

		startTime := time.Now()

		var messages *builtin.MessageAccumulator
		switch av {
		case actorstypes.Version8:
			messages, err = v8.CheckStateInvariants(actorTree, abi.ChainEpoch(height), actorCodeCids)
			if err != nil {
				return xerrors.Errorf("checking state invariants: %w", err)
			}
		case actorstypes.Version9:
			messages, err = v9.CheckStateInvariants(actorTree, abi.ChainEpoch(height), actorCodeCids)
			if err != nil {
				return xerrors.Errorf("checking state invariants: %w", err)
			}
		}

		fmt.Println("completed, took ", time.Since(startTime))

		for _, message := range messages.Messages() {
			fmt.Println("got the following error: ", message)
		}

		return nil
	},
}
