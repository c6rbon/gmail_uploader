# Simple mbox uploader for gmail

This tool inserts mail from an mbox into the credential owner's Gmail account.

*caveat emptor, gratis software*

**For current bugs and known issues, see [BUGS.md](BUGS.md)**

## Installation

* You need a GCP project and credential. Follow the instructions on the [Gmail API Quickstart](https://developers.google.com/gmail/api/quickstart/go) to get that set up, and install the golang libraries.
* Set up the mbox library
  * `go get -u github.com/sam-falvo/mbox.git`
  * go to the mbox directory, and `git apply` the included `sam-falvo-mbox.diff` (This is needed so we can get strictly ordered headers. My patch is not pretty.)
* `go build` etc.
* The first time you use it, will auth the needed OAuth scopes, and cache the token in `~/.credentials`.

## Notes

* Most logs are to stdout, prefixed by the mail file for easy grepping. FAILED imports attempt to provide mail file, message number and Message-id so you can try again, or debug the mbox.
* This generally relies on MBOXO parsing, but does a fallback check if there is a `Content-Length` field in the header. If there is, it will respect that.

## Flags

* `-n` which prevents actual message uploads, but will:

  * parse the mailbox
  * print each message as it is parsed
  * optionally output the base64 url encoded version of the message if you use `-print_encoded` (useful for manually using the [API via a browser](https://developers.google.com/gmail/api/v1/reference/users/messages/import))

* `-only_msgno` takes a series of comma deliniated message indices

* `-print_encoded` only makes sense when used it `-n`; it will print the base64 email instead of clear. Usedul for manual API testing.

* `-label` will attach a gmail label onto all uploaded messages

## Examples

### Single mailbox

Insert all mail:

```
./gmail_uploader -label imported my.mbox
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

## Thanks

* To the Gmail API folks, from whose quickstart the skeleton code was ripped with no shame.
* To MBOX format authors everywhere, for all of the [four MBOX varients](http://www.digitalpreservation.gov/formats/fdd/fdd000383.shtml) none of which are fully intercompatible.
* [Mark Crispen](https://en.wikipedia.org/wiki/Mark_Crispin), who is still my personal mail role-model.
