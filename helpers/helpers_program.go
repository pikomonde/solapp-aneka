package helpers

import (
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func CheckIfProgramExist(ctx context.Context, rpcCli *rpc.Client, programID string) error {
	programPubKey, err := solana.PublicKeyFromBase58(programID)
	if err != nil {
		return fmt.Errorf("invalid programID, err: ", err)
	}

	programAccountInfo, err := rpcCli.GetAccountInfo(
		ctx,
		programPubKey,
	)
	if err != nil {
		return fmt.Errorf("program not exist, err: ", err)
	}
	if !programAccountInfo.Value.Executable {
		return fmt.Errorf("program is not executable")
	}
	return nil
}
