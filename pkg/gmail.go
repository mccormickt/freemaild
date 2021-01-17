package freemaild

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/mail"
	"os"
	"strings"

	"github.com/pkg/errors"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Retrieve a token, saves the token, then returns the generated client.
func GetClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "/etc/freemaild/token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = tokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func tokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func GetOauthConfig(credentialFile string) (*oauth2.Config, error) {
	creds, err := ioutil.ReadFile(credentialFile)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(creds, gmail.GmailSendScope)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func getGmailClient(senderEmail, credentialFile string) (*gmail.Service, error) {
	ctx := context.Background()
	config, err := GetOauthConfig(credentialFile)
	if err != nil {
		return nil, err
	}
	client := GetClient(config)

	mailer, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, errors.Wrap(err, "failed to get gmail client with provided credentials")
	}
	return mailer, nil
}

func (f *Freemaild) gmailSMTPHandler(addr net.Addr, from string, to []string, data []byte) {
	msg, err := mail.ReadMessage(bytes.NewReader(data))
	if err != nil {
		log.Fatalf("error reading SMTP message: %+v", err)
	}

	// Organize headers
	emailFrom := fmt.Sprintf("From: %s\n", from)
	emailTo := fmt.Sprintf("To: %s\n", strings.Join(to, ","))
	subject := fmt.Sprintf("Subject: %s\n", msg.Header.Get("Subject"))
	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"utf-8\";\n\n"

	// Construct message from headers
	body, err := ioutil.ReadAll(msg.Body)
	if err != nil {
		log.Printf("error reading message body: %+v", err)
	}
	message := []byte(emailFrom + emailTo + subject + mime + "\n" + string(body))

	log.Printf("Forwarding mail recieved from %s for %s to Gmail", from, strings.Join(to, ","))

	// Send message to yourself
	gmailMessage := gmail.Message{
		Raw: base64.URLEncoding.EncodeToString(message),
	}
	_, err = f.mailer.Users.Messages.Send(f.config.RecipientEmail, &gmailMessage).Do()
	if err != nil {
		log.Printf("error sending email via Gmail: %+v", err)
	}
}

// TODO: Add webhook gmail handler
//func (f *Freemaild) gmailHTTPHandler() error { return }
