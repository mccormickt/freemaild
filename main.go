package main

import (
	freemaild "github.com/jan0ski/freemaild/pkg"
)

func main() {
	config := freemaild.Config{
		CredentialSecret: "",
		CredentialFile:   "credentials.json",
		Address:          "127.0.0.1",
		Port:             2525,
		TLS:              nil,
	}

	server := freemaild.New(&config)
	server.Run()
}
