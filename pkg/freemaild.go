package freemaild

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/mhale/smtpd"
	"google.golang.org/api/gmail/v1"
)

const versionString = "v0.1"

// Config for Freemaild
type Config struct {
	CredentialFile string
	SenderEmail    string
	RecipientEmail string
	Address        string
	Port           int64
	TLS            map[string]string
}

// Freemaild is an SMTP to Gmail Relay
type Freemaild struct {
	config *Config
	mailer *gmail.Service
}

// New creates a new instance of Freemaild
func New(cfg *Config) (*Freemaild, error) {
	gmailClient, err := getGmailClient(cfg.SenderEmail, cfg.CredentialFile)
	if err != nil {
		return nil, err
	}
	gmailClient.UserAgent = "Freemaild " + versionString

	client := Freemaild{config: cfg, mailer: gmailClient}
	return &client, nil
}

// Run the daemon
func (f *Freemaild) Run() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	// TODO: Add switch for different mail handlers?

	// start SMTP handler
	smtpAddress := fmt.Sprintf("%s:%d", f.config.Address, f.config.Port)
	log.Printf("Listening for mail on %s:%d", f.config.Address, f.config.Port)
	if f.config.TLS != nil {
		go smtpd.ListenAndServeTLS(smtpAddress, f.config.TLS["cert"], f.config.TLS["key"], f.gmailSMTPHandler, "freemaild", "")
	} else {
		go smtpd.ListenAndServe(smtpAddress, f.gmailSMTPHandler, "freemaild", "")
	}

	<-sig
	return
}
