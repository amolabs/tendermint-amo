package evidence

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	dbm "github.com/amolabs/tendermint-amo/libs/db"
	sm "github.com/amolabs/tendermint-amo/state"
	"github.com/amolabs/tendermint-amo/types"
	tmtime "github.com/amolabs/tendermint-amo/types/time"
)

func TestMain(m *testing.M) {
	types.RegisterMockEvidences(cdc)

	code := m.Run()
	os.Exit(code)
}

func initializeValidatorState(valAddr []byte, height int64) dbm.DB {
	stateDB := dbm.NewMemDB()

	// create validator set and state
	valSet := &types.ValidatorSet{
		Validators: []*types.Validator{
			{Address: valAddr},
		},
	}
	state := sm.State{
		LastBlockHeight:             0,
		LastBlockTime:               tmtime.Now(),
		Validators:                  valSet,
		NextValidators:              valSet.CopyIncrementProposerPriority(1),
		LastHeightValidatorsChanged: 1,
		ConsensusParams: types.ConsensusParams{
			Evidence: types.EvidenceParams{
				MaxAge: 1000000,
			},
		},
	}

	// save all states up to height
	for i := int64(0); i < height; i++ {
		state.LastBlockHeight = i
		sm.SaveState(stateDB, state)
	}

	return stateDB
}

func TestEvidencePool(t *testing.T) {

	valAddr := []byte("val1")
	height := int64(5)
	stateDB := initializeValidatorState(valAddr, height)
	store := NewEvidenceStore(dbm.NewMemDB())
	pool := NewEvidencePool(stateDB, store)

	goodEvidence := types.NewMockGoodEvidence(height, 0, valAddr)
	badEvidence := types.MockBadEvidence{goodEvidence}

	// bad evidence
	err := pool.AddEvidence(badEvidence)
	assert.NotNil(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		<-pool.EvidenceWaitChan()
		wg.Done()
	}()

	err = pool.AddEvidence(goodEvidence)
	assert.Nil(t, err)
	wg.Wait()

	assert.Equal(t, 1, pool.evidenceList.Len())

	// if we send it again, it shouldnt change the size
	err = pool.AddEvidence(goodEvidence)
	assert.Nil(t, err)
	assert.Equal(t, 1, pool.evidenceList.Len())
}
