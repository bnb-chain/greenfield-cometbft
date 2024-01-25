package state

import (
	"errors"
	"fmt"

	cmtstate "github.com/cometbft/cometbft/proto/tendermint/state"
	cmtversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"github.com/cometbft/cometbft/types"
	"github.com/cometbft/cometbft/version"
)

// Rollback overwrites the current CometBFT state (height n) with the most
// recent previous state (height n - 1).
// Note that this function does not affect application state.
func Rollback(bs BlockStore, ss Store, removeBlock bool, rollbackBlocks int64) (int64, []byte, error) {
	invalidState, err := ss.Load()
	if err != nil {
		return -1, nil, err
	}
	if invalidState.IsEmpty() {
		return -1, nil, errors.New("no state found")
	}

	height := bs.Height()

	// NOTE: persistence of state and blocks don't happen atomically. Therefore it is possible that
	// when the user stopped the node the state wasn't updated but the blockstore was. Discard the
	// pending block before continuing.
	if height == invalidState.LastBlockHeight+1 {
		if removeBlock {
			if err := bs.DeleteLatestBlock(); err != nil {
				return -1, nil, fmt.Errorf("failed to remove final block from blockstore: %w", err)
			}
		}
		rollbackBlocks--
		if rollbackBlocks == 0 {
			return invalidState.LastBlockHeight, invalidState.AppHash, nil
		}
	} else if height != invalidState.LastBlockHeight {
		return -1, nil, fmt.Errorf("statestore height (%d) is not one below or equal to blockstore height (%d)",
			invalidState.LastBlockHeight, height)
	}

	// state store height is equal to blockstore height. We're good to proceed with rolling back state
	rollbackHeight := invalidState.LastBlockHeight - rollbackBlocks
	rollbackBlock := bs.LoadBlockMeta(rollbackHeight)
	if rollbackBlock == nil {
		return -1, nil, fmt.Errorf("block at height %d not found", rollbackHeight)
	}

	// We also need to retrieve the latest block because the app hash and last
	// results hash is only agreed upon in the following block.
	latestBlock := bs.LoadBlockMeta(rollbackHeight + 1)
	if latestBlock == nil {
		return -1, nil, fmt.Errorf("block at height %d not found", invalidState.LastBlockHeight)
	}

	previousLastValidatorSet, err := ss.LoadValidators(rollbackHeight)
	if err != nil {
		return -1, nil, err
	}

	previousParams, err := ss.LoadConsensusParams(rollbackHeight + 1)
	if err != nil {
		return -1, nil, err
	}

	nextHeight := rollbackHeight + 1
	lastHeightValidatorsChanged, err := ss.LoadLastHeightValidatorsChanged(nextHeight + 1)
	if err != nil {
		return -1, nil, err
	}

	// this can only happen if the validator set changed since the last block
	if lastHeightValidatorsChanged > nextHeight {
		lastHeightValidatorsChanged = nextHeight + 1
	}

	//paramsChangeHeight := invalidState.LastHeightConsensusParamsChanged
	paramsChangeHeight, err := ss.LoadLastHeightConsensusParamsChanged(nextHeight)
	if err != nil {
		return -1, nil, err
	}
	// this can only happen if params changed from the last block
	if paramsChangeHeight > rollbackHeight {
		paramsChangeHeight = rollbackHeight + 1
	}

	var validators *types.ValidatorSet
	var nextValidators *types.ValidatorSet
	if rollbackBlocks == 1 {
		validators = invalidState.LastValidators
		nextValidators = invalidState.Validators
	} else {
		validators, err = ss.LoadValidators(rollbackHeight + 1)
		if err != nil {
			return -1, nil, err
		}
		nextValidators, err = ss.LoadValidators(rollbackHeight + 2)
		if err != nil {
			return -1, nil, err
		}
	}

	// build the new state from the old state and the prior block
	rolledBackState := State{
		Version: cmtstate.Version{
			Consensus: cmtversion.Consensus{
				Block: version.BlockProtocol,
				App:   previousParams.Version.App,
			},
			Software: version.TMCoreSemVer,
		},
		// immutable fields
		ChainID:                     invalidState.ChainID,
		InitialHeight:               invalidState.InitialHeight,
		LastBlockHeight:             rollbackBlock.Header.Height,
		LastBlockID:                 rollbackBlock.BlockID,
		LastBlockTime:               rollbackBlock.Header.Time,
		NextValidators:              nextValidators,
		Validators:                  validators,
		LastValidators:              previousLastValidatorSet,
		LastHeightValidatorsChanged: lastHeightValidatorsChanged,

		ConsensusParams:                  previousParams,
		LastHeightConsensusParamsChanged: paramsChangeHeight,

		LastResultsHash: latestBlock.Header.LastResultsHash,
		AppHash:         latestBlock.Header.AppHash,

		LastRandaoMix: rollbackBlock.Header.RandaoMix,
	}

	// persist the new state. This overrides the invalid one. NOTE: this will also
	// persist the validator set and consensus params over the existing structures,
	// but both should be the same
	if err := ss.Save(rolledBackState); err != nil {
		return -1, nil, fmt.Errorf("failed to save rolled back state: %w", err)
	}

	// If removeBlock is true then also remove the block associated with the previous state.
	// This will mean both the last state and last block height is equal to n - 1
	if removeBlock {
		if err := bs.DeleteLatestBlocks(uint64(rollbackBlocks)); err != nil {
			return -1, nil, fmt.Errorf("failed to remove final block from blockstore: %w", err)
		}
	}

	return rolledBackState.LastBlockHeight, rolledBackState.AppHash, nil
}
