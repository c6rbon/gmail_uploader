# Simple mbox uploader for gmail

This expects a local `client_secret.json` in the pwd, and inserts the mail into the credential owner's account.

The first time you use it, will auth the needed OAuth scopes, and cache the token in `~/.credentials`.

Most logs are to stdout, prefixed by the mail file for easy grepping. FAILED imports attempt to provide mail file, message number and Message-id so you can try again, or debug the mbox.

## Testing

This supports a `-n` argument, which prevents actual message uploads, but will:

* parse the mailbox
* print each message as it is parsed
* output the base64 url encoded version of the message (useful for manually using the API via a browser)

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
