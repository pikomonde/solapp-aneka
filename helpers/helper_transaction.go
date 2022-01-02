package helpers

import (
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

func NewTransactionAndSignAndSendAndConfirm(
	ctx context.Context,
	rpcCli *rpc.Client,
	wsCli *ws.Client,
	instructions []solana.Instruction,
	signers []solana.PrivateKey,
	skipPreflight bool, // if true, skip the preflight transaction checks (default: false)
	preflightCommitment rpc.CommitmentType, // optional; Commitment level to use for preflight (default: "confirmed").
	commitment rpc.CommitmentType, // optional; Commitment level to use for preflight (default: "confirmed").
	opts ...solana.TransactionOption,
) (*solana.Transaction, error) {

	// create recent blockhash
	recent, err := rpcCli.GetRecentBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, err
	}

	// create transaction
	tx, err := solana.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	// sign transaction
	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			for _, signer := range signers {
				if signer.PublicKey().Equals(key) {
					return &signer
				}
			}

			return nil
		},
	)
	if err != nil {
		return tx, err
	}

	// send and confirm transaction
	_, err = SendAndConfirmTransactionWithOpts(
		ctx,
		rpcCli,
		wsCli,
		tx,
		skipPreflight,
		preflightCommitment,
		commitment,
	)
	if err != nil {
		return tx, err
	}

	return tx, nil
}

func SendAndConfirmTransactionWithOpts(
	ctx context.Context,
	rpcClient *rpc.Client,
	wsClient *ws.Client,
	transaction *solana.Transaction,
	skipPreflight bool, // if true, skip the preflight transaction checks (default: false)
	preflightCommitment rpc.CommitmentType, // optional; Commitment level to use for preflight (default: "confirmed").
	commitment rpc.CommitmentType, // optional; Commitment level to use for preflight (default: "confirmed").
) (signature solana.Signature, err error) {

	if preflightCommitment == "" {
		preflightCommitment = rpc.CommitmentConfirmed
	}
	if commitment == "" {
		commitment = rpc.CommitmentConfirmed
	}

	sig, err := rpcClient.SendTransactionWithOpts(
		ctx,
		transaction,
		skipPreflight,
		preflightCommitment,
	)
	if err != nil {
		return sig, err
	}

	sub, err := wsClient.SignatureSubscribe(
		sig,
		commitment,
	)
	if err != nil {
		return sig, err
	}
	defer sub.Unsubscribe()

	for {
		got, err := sub.Recv()
		if err != nil {
			return sig, err
		}
		if got.Value.Err != nil {
			return sig, fmt.Errorf("transaction confirmation failed: %v", got.Value.Err)
		} else {
			return sig, nil
		}
	}
}
