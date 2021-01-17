# freemaild
SMTP relay for sending emails via Gmail's API.


## Usage
1. Create a new project on Google Cloud Platform (optional) and activate the Gmail API, using the account you wish Freemaild to send emails from.
2. Create an OAuth 2.0 Client ID for Freemaild to use, save the credential file as `freemaild.json`.
3. Grab an OAuth token for your gmail account. It will be saved to `/etc/freemaild/token.json` by default.
```bash
$ ./freemaild init
```
A link will be displayed on the console. Navigate to it in a browser, log into your desired Google account, and accept the dialogue after allowing your Freemaild instance to send emails on your behalf.  Enter the token given from Google back into the console, creating your token.

4. Run the server, specifying server address, listen port, and app credential file *(from step 2)* path via environment variables.
```bash
$ FREEMAILD_ADDRESS=127.0.0.1 FREEMAILD_PORT=2025 ./freemaild
```
5. Send an email via SMTP to the server, it should forward it out to the recipient via your Gmail account!

### Docker
A docker image is available and can be used as such:
```bash
$ docker run -it -v /etc/freemaild:/etc/freemaild ghcr.io/jan0ski/freemaild:latest [init]
```