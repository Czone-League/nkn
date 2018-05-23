package voting

import (
	"sync"

	. "github.com/nknorg/nkn/common"
	"github.com/nknorg/nkn/core/ledger"
	"github.com/nknorg/nkn/core/transaction"
	"github.com/nknorg/nkn/por"
)

type SigChainVoting struct {
	sync.RWMutex
	pstate         map[Uint256]*State            // consensus state for proposer
	vstate         map[uint64]map[Uint256]*State // consensus state for voter
	height         uint32                        // voting height
	porServer      *por.PorServer                // signature chain service provider
	pool           *SigChainVotingPool           // signature chain voting pool
	confirmingHash Uint256                       // signature chain hash in process
	txnCollector   *transaction.TxnCollector     // transaction pool
}

func NewSigChainVoting(totalWeight int, porServer *por.PorServer, txnCollector *transaction.TxnCollector) *SigChainVoting {
	sigChainVoting := &SigChainVoting{
		pstate:       make(map[Uint256]*State),
		vstate:       make(map[uint64]map[Uint256]*State),
		height:       ledger.DefaultLedger.Store.GetHeight() + 2,
		porServer:    porServer,
		pool:         NewSigChainVotingPool(totalWeight),
		txnCollector: txnCollector,
	}

	return sigChainVoting
}

func (scv *SigChainVoting) SetProposerState(hash Uint256, s State) {
	scv.Lock()
	defer scv.Unlock()

	if _, ok := scv.pstate[hash]; !ok {
		scv.pstate[hash] = new(State)
	}
	scv.pstate[hash].SetBit(s)
}

func (scv *SigChainVoting) HasProposerState(hash Uint256, state State) bool {
	scv.RLock()
	defer scv.RUnlock()

	if v, ok := scv.pstate[hash]; !ok || v == nil {
		return false
	} else {
		if v.HasBit(state) {
			return true
		}
		return false
	}
}

func (scv *SigChainVoting) SetVoterState(id uint64, hash Uint256, s State) {
	scv.Lock()
	defer scv.Unlock()

	if _, ok := scv.vstate[id]; !ok {
		scv.vstate[id] = make(map[Uint256]*State)
	}
	if _, ok := scv.vstate[id][hash]; !ok {
		scv.vstate[id][hash] = new(State)
	}
	scv.vstate[id][hash].SetBit(s)
}

func (scv *SigChainVoting) HasVoterState(id uint64, hash Uint256, state State) bool {
	scv.RLock()
	defer scv.RUnlock()

	if _, ok := scv.vstate[id]; !ok {
		return false
	} else {
		if v, ok := scv.vstate[id][hash]; !ok || v == nil {
			return false
		} else {
			if v.HasBit(state) {
				return true
			}
			return false
		}
	}
}

func (scv *SigChainVoting) SetVotingHeight(height uint32) {
	scv.height = height
}

func (scv *SigChainVoting) UpdateVotingHeight() {
	scv.height = ledger.DefaultLedger.Store.GetHeight() + 2
}

func (scv *SigChainVoting) GetVotingHeight() uint32 {
	return scv.height
}

func (scv *SigChainVoting) SetConfirmingHash(hash Uint256) {
	scv.confirmingHash = hash
}

func (scv *SigChainVoting) GetConfirmingHash() Uint256 {
	return scv.confirmingHash
}

func (scv *SigChainVoting) GetBestVotingContent(height uint32) (VotingContent, error) {
	sigChain, err := scv.porServer.GetMinSigChain(height)
	if err != nil {
		return nil, err
	}
	txnHash, _ := scv.porServer.IsSigChainExist(sigChain.Hash(), height)
	txn := scv.txnCollector.GetTransaction(*txnHash)

	return txn, nil
}

func (scv *SigChainVoting) GetWorseVotingContent(height uint32) (VotingContent, error) {
	return nil, nil
}

func (scv *SigChainVoting) GetVotingContent(hash Uint256, height uint32) (VotingContent, error) {
	sigChain, err := scv.porServer.GetSigChain(height, hash)
	if err != nil {
		return nil, err
	}
	txnHash, _ := scv.porServer.IsSigChainExist(sigChain.Hash(), height)
	txn := scv.txnCollector.GetTransaction(*txnHash)

	return txn, nil
}

func (scv *SigChainVoting) VotingType() VotingContentType {
	return SigChainTxnVote
}

func (scv *SigChainVoting) Preparing(content VotingContent) error {
	return nil
}

func (scv *SigChainVoting) Exist(hash Uint256, height uint32) bool {
	ret := scv.txnCollector.GetTransaction(hash)
	if ret == nil {
		return false
	}

	return true
}

func (scv *SigChainVoting) GetVotingPool() VotingPool {
	return scv.pool
}

func (scv *SigChainVoting) DumpState(hash Uint256, desc string, verbose bool) {
}