package main

import (
	"context"
	"log"

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
	cliHello, err := clientAnekaHello.InitClient(ctx, clientAnekaHello.ClientOption{
		PayerAccount: payerAcc,
	})
	if err != nil {
		panic(err)
	}

	// Say hello
	err = cliHello.SayHello()
	if err != nil {
		panic(err)
	}

	// Check how many has been greeted
	greetData, err := cliHello.ReportGreetings()
	if err != nil {
		panic(err)
	}

	log.Printf("Account %s greets %d times\n", payerAcc.PublicKey(), greetData.Counter)
	log.Println("Success")
}
