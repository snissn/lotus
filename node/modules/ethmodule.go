package modules

import (
	"context"

	"go.uber.org/fx"

	"github.com/filecoin-project/lotus/chain/events"
	"github.com/filecoin-project/lotus/chain/events/filter"
	"github.com/filecoin-project/lotus/chain/messagepool"
	"github.com/filecoin-project/lotus/chain/stmgr"
	"github.com/filecoin-project/lotus/chain/store"
	"github.com/filecoin-project/lotus/node/config"
	"github.com/filecoin-project/lotus/node/impl/full"
	"github.com/filecoin-project/lotus/node/modules/helpers"
)

func EthModuleAPI(cfg config.EthTxHashConfig) func(helpers.MetricsCtx, fx.Lifecycle, *store.ChainStore, *stmgr.StateManager, EventAPI, *messagepool.MessagePool, full.StateAPI, full.ChainAPI, full.MpoolAPI) (*full.EthModule, error) {
	return func(mctx helpers.MetricsCtx, lc fx.Lifecycle, cs *store.ChainStore, sm *stmgr.StateManager, evapi EventAPI, mp *messagepool.MessagePool, stateapi full.StateAPI, chainapi full.ChainAPI, mpoolapi full.MpoolAPI) (*full.EthModule, error) {
		em := &full.EthModule{
			Chain:        cs,
			Mpool:        mp,
			StateManager: sm,
			ChainAPI:     chainapi,
			MpoolAPI:     mpoolapi,
			StateAPI:     stateapi,
		}

		if !cfg.EnableEthHashToFilecoinCidMapping {
			// mapping functionality disabled. Nothing to do here
			return em, nil
		}

		eventIndex, err := filter.NewEventIndex(cfg.TransactionHashLookupDatabasePath)
		if err != nil {
			return nil, err
		}

		lc.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				return eventIndex.Close()
			},
		})

		ethTxHashManager := full.EthTxHashManager{
			StateAPI:   stateapi,
			EventIndex: eventIndex,
		}

		em.EthTxHashManager = &ethTxHashManager

		const ChainHeadConfidence = 1

		ctx := helpers.LifecycleCtx(mctx, lc)
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				ev, err := events.NewEventsWithConfidence(ctx, &evapi, ChainHeadConfidence)
				if err != nil {
					return err
				}

				// Tipset listener
				_ = ev.Observe(&ethTxHashManager)

				ch, err := mp.Updates(ctx)
				if err != nil {
					return err
				}
				go full.WaitForMpoolUpdates(ctx, ch, &ethTxHashManager)

				return nil
			},
		})

		return em, nil
	}
}
