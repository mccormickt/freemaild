package freemaild

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/mail"
	"os"
	"os/signal"
	"strings"

	"github.com/mhale/smtpd"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

// Config for Freemaild
type Config struct {
	CredentialSecret string
	CredentialFile   string
	Address          string
	Port             int32
	TLS              map[string]string
}

// Freemaild is an SMTP to Gmail Relay
type Freemaild struct {
	config *Config
	mailer *gmail.Service
}

// New creates a new instance of Freemaild
func New(cfg *Config) *Freemaild {
	// TODO: add option for to k8s secret
	var b []byte
	b, err := ioutil.ReadFile(cfg.CredentialFile)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	gmailClient := getGmailClient(config)

	mailer, err := gmail.New(gmailClient)
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	client := Freemaild{config: cfg, mailer: mailer}
	return &client
}

// Run the daemon
func (f *Freemaild) Run() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	// TODO: Add switch for mail handlers?

	// start SMTP handler
	smtpAddress := fmt.Sprintf("%s:%d", f.config.Address, f.config.Port)
	log.Printf("Listening for mail on %s:%d", f.config.Address, f.config.Port)
	if f.config.TLS != nil {
		smtpd.ListenAndServeTLS(smtpAddress, f.config.TLS["cert"], f.config.TLS["key"], f.gmailSMTPHandler, "freemaild", "")
	} else {
		smtpd.ListenAndServe(smtpAddress, f.gmailSMTPHandler, "freemaild", "")
	}
	<-sig
	return
}

func (f *Freemaild) gmailSMTPHandler(origin net.Addr, from string, to []string, data []byte) {
	msg, err := mail.ReadMessage(bytes.NewReader(data))
	if err != nil {
		log.Fatalf("error parsing message: %+v", err)
	}

	// Organize headers
	emailFrom := fmt.Sprintf("From: %s\n", from)
	emailTo := fmt.Sprintf("To: %s\n", strings.Join(to, ","))
	subject := fmt.Sprintf("Subject: %s\n", msg.Header.Get("Subject"))
	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"utf-8\";\n\n"

	// Construct message from headers
	body, err := ioutil.ReadAll(msg.Body)
	if err != nil {
		log.Fatalf("error parsing message body: %+v", err)
	}
	message := []byte(emailFrom + emailTo + subject + mime + "\n" + string(body))

	log.Printf("Forwarding mail recieved from %s for %s to Gmail", from, strings.Join(to, ","))

	// Send message to yourself
	gmailMessage := gmail.Message{
		Raw: base64.URLEncoding.EncodeToString(message),
	}
	_, err = f.mailer.Users.Messages.Send("me", &gmailMessage).Do()
	if err != nil {
		log.Fatalf("unable to send mail to gmail: %+v", err)
	}
}
