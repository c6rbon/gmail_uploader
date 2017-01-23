# Simple mbox uploader for gmail

Example:

```bash
echo /Users/bhk/seraph/jargon/inky/jargon/mail/* | xargs -P10 -n 1 ./gmail_uploader | tee mail_log
```

Then clean up:

```bash
grep FAILED mail_log | sort -u | ./find_remaining.pl > replay.args
cat replay.args | xargs -P10 -n 2 ./gmail_uploader -only_msgno | tee new_mail_log
```