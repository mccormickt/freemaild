package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	freemaild "github.com/jan0ski/freemaild/pkg"
)

var credentialFile, senderEmail, recipientEmail, address string

func main() {
	if credentialFile = os.Getenv("FREEMAILD_CRED_FILE"); credentialFile == "" {
		credentialFile = "/etc/freemaild/freemaild.json"
		err := os.MkdirAll(filepath.Dir(credentialFile), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
	if senderEmail = os.Getenv("FREEMAILD_SENDER_EMAIL"); senderEmail == "" {
		senderEmail = "root@freemaild.lan"
	}
	if recipientEmail = os.Getenv("FREEMAILD_RECIPIENT_EMAIL"); recipientEmail == "" {
		recipientEmail = "me"
	}
	if address = os.Getenv("FREEMAILD_ADDRESS"); address == "" {
		address = "0.0.0.0"
	}
	port, err := strconv.ParseInt(os.Getenv("FREEMAILD_PORT"), 10, 64)
	if err != nil {
		port = 25
		log.Printf("couldn't assign specified listen port, using default: %d", port)
	}

	if len(os.Args) > 1 && os.Args[1] == "init" {
		config, err := freemaild.GetOauthConfig(credentialFile)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		freemaild.GetClient(config)

		f, err := os.Open(filepath.Dir(credentialFile) + "/token.json")
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		defer f.Close()

		fmt.Println("Freemaild OAuth token saved to 'token.json'.  Keep this file safe.")
		os.Exit(0)
	}

	// TODO: Support TLS certificate arguments
	config := freemaild.Config{
		CredentialFile: credentialFile,
		SenderEmail:    senderEmail,
		RecipientEmail: recipientEmail,
		Address:        address,
		Port:           port,
		TLS:            nil,
	}

	server, err := freemaild.New(&config)
	if err != nil {
		log.Fatalf("error creating freemaild client: %+v", err)
		os.Exit(1)
	}
	server.Run()
}
