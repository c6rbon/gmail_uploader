# Simple mbox uploader for gmail

This tool inserts mail from an mbox into the credential owner's Gmail account.

Most logs are to stdout, prefixed by the mail file for easy grepping. FAILED imports attempt to provide mail file, message number and Message-id so you can try again, or debug the mbox.

**For current bugs and known issues, see [BUGS.md](BUGS.md)**

## Installation

* You need a GCP project and credential. Follow the instructions on the [Gmail API Quickstart](https://developers.google.com/gmail/api/quickstart/go) to get that set up, and install the golang libraries.
* Set up the mbox library
  * `go get -u github.com/sam-falvo/mbox.git`
  * go to the mbox directory, and `git apply` the included `sam-falvo-mbox.diff`
* `go build` etc.
* The first time you use it, will auth the needed OAuth scopes, and cache the token in `~/.credentials`.

## Testing

This supports a `-n` argument, which prevents actual message uploads, but will:

* parse the mailbox
* print each message as it is parsed
* output the base64 url encoded version of the message (useful for manually using the [API via a browser](https://developers.google.com/gmail/api/v1/reference/users/messages/import))

## Examples

### Single mailbox

Insert all mail:

```
./gmail_uploader my.mbox
```

Insert just message numbers 7 and 9:

```
./gmail_uploader -only_msgno 7,9 my.mbox
```

### Parallel use

```
echo /Users/bhk/seraph/jargon/inky/jargon/mail/* | xargs -P10 -n 1 ./gmail_uploader | tee mail_log
```

Then clean up:

```
grep FAILED mail_log | sort -u | ./find_remaining.pl > replay.args
cat replay.args | xargs -P10 -n 2 ./gmail_uploader -only_msgno | tee new_mail_log
```
