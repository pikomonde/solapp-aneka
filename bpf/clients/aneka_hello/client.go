package aneka_hello

import (
	"bytes"
	"context"
	"fmt"
	"log"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/pikomonde/solapp-aneka/helpers"
)

type Client struct {
	ctx                   context.Context
	rpcClient             *rpc.Client
	wsClient              *ws.Client
	programPubKey         solana.PublicKey
	payerAccount          solana.Wallet
	payerSubAccountPubKey solana.PublicKey
}

type ClientOption struct {
	RPCURL       string // optional, leave it empty to use local
	WSURL        string // optional, leave it empty to use local
	PayerAccount *solana.Wallet
}

const (
	programID         = "J3aRWdbSPs7BLKkmyXMo8LEdhbKLkXZDLjbVUQUuA9Lr"
	seedAnekaHello    = "hello"
	defaultCommitment = rpc.CommitmentConfirmed
)

func InitClient(ctx context.Context, opt ClientOption) (*Client, error) {
	// Establish connection to the cluster
	if opt.RPCURL == "" {
		opt.RPCURL = rpc.LocalNet_RPC
	}
	if opt.WSURL == "" {
		opt.WSURL = rpc.LocalNet_WS
	}

	rpcClient := rpc.New(opt.RPCURL)
	version, err := rpcClient.GetVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to rpc Client, err: %v", err)
	}

	wsClient, err := ws.Connect(ctx, opt.WSURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to websocket Client, err: %v", err)
	}

	log.Println("Connection to cluster established:", opt.RPCURL, opt.WSURL, version)

	// Create client
	programPubKey, err := solana.PublicKeyFromBase58(programID)
	if err != nil {
		return nil, fmt.Errorf("Invalid ProgramID: %s, err: %v", programID, err)
	}
	if opt.PayerAccount == nil {
		return nil, fmt.Errorf("PayerAccount (%s) should not null", opt.PayerAccount)
	}
	cli := &Client{
		ctx:           ctx,
		rpcClient:     rpcClient,
		wsClient:      wsClient,
		programPubKey: programPubKey,
		payerAccount:  *opt.PayerAccount,
	}

	// Check if program exist
	if err = helpers.CheckIfProgramExist(cli.ctx, cli.rpcClient, programID); err != nil {
		return nil, fmt.Errorf("Failed on CheckIfProgramExist, err:", err)
	}

	// Create sub account if not provided (and not exist)
	if err = cli.createSubAccountIfNotExist(); err != nil {
		return nil, fmt.Errorf("Failed on createSubAccountIfNotExist, err:", err)
	}

	return cli, nil
}

func (cli *Client) createSubAccountIfNotExist() error {
	var err error

	// Create SubAccount PubKey
	cli.payerSubAccountPubKey, err = solana.CreateWithSeed(
		cli.payerAccount.PublicKey(),
		seedAnekaHello,
		cli.programPubKey,
	)
	if err != nil {
		return fmt.Errorf("failed to CreateWithSeed, err: ", err)
	}

	// Create SubAccount if not exist
	_, err = cli.rpcClient.GetAccountInfo(cli.ctx, cli.payerSubAccountPubKey)
	if err != nil {
		log.Printf("Payer sub account %s does not exist, getting this error message (%v), creating...\n",
			cli.payerSubAccountPubKey,
			err,
		)

		// Calculating data size and lamports needed
		var buff bytes.Buffer
		bin.NewBorshEncoder(&buff).Encode(GreetingAccount{})
		dataSize := uint64(buff.Len())

		lamports, err := cli.rpcClient.GetMinimumBalanceForRentExemption(
			cli.ctx,
			dataSize,
			defaultCommitment,
		)
		if err != nil {
			return fmt.Errorf("failed to GetMinimumBalanceForRentExemption, err: ", err)
		}

		// TODO: check payer balance

		// Set instruction
		instruction, err := system.NewCreateAccountWithSeedInstruction(
			cli.payerAccount.PublicKey(),
			seedAnekaHello,
			lamports,
			dataSize,
			cli.programPubKey,
			cli.payerAccount.PublicKey(),
			cli.payerSubAccountPubKey,
			cli.payerAccount.PublicKey(),
		).ValidateAndBuild()
		if err != nil {
			log.Fatalln("Failed building NewCreateAccountInstruction instruction", err)
		}

		// Set transaction, sign, send, and confirm
		transaction, err := helpers.NewTransactionAndSignAndSendAndConfirm(
			cli.ctx, cli.rpcClient, cli.wsClient,
			[]solana.Instruction{
				instruction,
			},
			[]solana.PrivateKey{cli.payerAccount.PrivateKey},
			false, defaultCommitment, defaultCommitment,
			solana.TransactionPayer(cli.payerAccount.PublicKey()),
		)
		if err != nil {
			log.Fatalln("Failed to NewTransactionAndSignAndSendAndConfirm GetAccountInfo", err)
		}

		_ = transaction
		// transaction.EncodeTree(text.NewTreeEncoder(os.Stdout, "Create Account"))
	}
	return nil
}

func (cli *Client) SayHello() error {

	instruction := solana.NewInstruction(
		cli.programPubKey,
		solana.AccountMetaSlice{&solana.AccountMeta{
			PublicKey: cli.payerSubAccountPubKey,
			IsSigner:  false, IsWritable: true,
		}},
		[]byte{},
	)

	transaction, err := helpers.NewTransactionAndSignAndSendAndConfirm(
		cli.ctx, cli.rpcClient, cli.wsClient,
		[]solana.Instruction{
			instruction,
		},
		[]solana.PrivateKey{cli.payerAccount.PrivateKey},
		false, defaultCommitment, defaultCommitment,
		solana.TransactionPayer(cli.payerAccount.PublicKey()),
	)
	if err != nil {
		log.Fatalln("Failed to NewTransactionAndSignAndSendAndConfirm SayHello", err)
	}

	_ = transaction
	// transaction.EncodeTree(text.NewTreeEncoder(os.Stdout, "Create Account"))

	return nil
}

func (cli *Client) ReportGreetings() error {
	payerSubAccountInfo, err := cli.rpcClient.GetAccountInfoWithOpts(cli.ctx, cli.payerSubAccountPubKey, &rpc.GetAccountInfoOpts{Encoding: solana.EncodingBase64, Commitment: rpc.CommitmentConfirmed})
	// payerSubAccountInfo, err := cli.rpcClient.GetAccountInfo(cli.ctx, cli.payerSubAccountPubKey)
	if err != nil {
		log.Fatalln("Failed to GetAccountInfo SayHello", err)
	}

	var payerSubAccountInfoData GreetingAccount
	bin.NewBorshDecoder(payerSubAccountInfo.Value.Data.GetBinary()).Decode(&payerSubAccountInfoData)
	fmt.Println("-----> payerSubAccountInfoData", payerSubAccountInfoData)
	return nil
}

// export async function reportGreetings(): Promise<void> {
// 	const accountInfo = await connection.getAccountInfo(greetedPubkey);
// 	if (accountInfo === null) {
// 	  throw 'Error: cannot find the greeted account';
// 	}
// 	const greeting = borsh.deserialize(
// 	  GreetingSchema,
// 	  GreetingAccount,
// 	  accountInfo.data,
// 	);
// 	console.log(
// 	  greetedPubkey.toBase58(),
// 	  'has been greeted',
// 	  greeting.counter,
// 	  'time(s)',
// 	);
//   }
