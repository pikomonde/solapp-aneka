package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gagliardetto/solana-go"
	clientAnekaHello "github.com/pikomonde/solapp-aneka/bpf/clients/aneka_hello"
)

func main() {
	ctx := context.Background()

	log.Println("Let's say hello to a Solana account...")

	// Define accounts
	payerPrivateKey, _ := solana.PrivateKeyFromSolanaKeygenFile("/home/zenga/.config/solana/id01.json")
	payerAcc, _ := solana.WalletFromPrivateKeyBase58(payerPrivateKey.String())

	// Client
	cliHellp, err := clientAnekaHello.InitClient(ctx, clientAnekaHello.ClientOption{
		PayerAccount: payerAcc,
	})
	if err != nil {
		panic(err)
	}

	// Say hello
	now := time.Now()
	cliHellp.SayHello()
	fmt.Println(time.Since(now))

	// Check how many has been greeted
	cliHellp.ReportGreetings()

	log.Println("Success")
}
