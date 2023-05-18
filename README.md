## Mailout

Simple mailer for SendGrid

### Configure

Edit the existing config.json or create another one.
It's pretty self-explanatory

    {
        "from": "mail@example.com",
        "fromName": "SenderName",
        "subject": "Test!",
        "attachments": [
            {"type": "image/png", "fileName": "attachments/some_pic.png"},
            {"type": "application/pdf", "fileName": "attachments/some_doc.pdf"}
        ],
        "useBCC": false,
        "bccPerEmail": 1,
        "recipientsFile": "recipients.txt",
        "htmlEmailFile": "email.html"
    }

`from` the address the mail is from

`fromName` the name the mail is from

`subject` the subject

`attachments` array of files to attach. Does not need to be included in the config, but if it is,
enter the correct mime type for the file in `type` and the filename relative
to the binary in `fileName`

`useBCC` if there are lots of emails to send you can set this to true
and it will send `bccPerEmail` number of emails in the BCC field for each mail.
SMTP standard (and SendGrid) does not allow a mail to have only
BCC addresses though, so there needs to be at least one address in the
to-field. This will be set to the value in `from`. It is generally not recommended
to use this feature as it might decrease deliverability. If this field is set to
false the `bccPerEmail` will have no effect and only one mail will be sent in each
call to SendGrid.

`recipientsFile` newline separated text file with one email per line

`htmlEmailFile` the email to send. There is currently no support for sending text-based emails or having text fallback.


### Command line options

`-config <file>` use this config file. Defaults to config.json

`-v` verbose mode, output the emails being send and show more
information about any errors the only the summary at the end.

`-help` show help

### Compile

    go mod vendor
    go build

Or use the precompiled binaries in /bin

### Customize the email

Edit `email.html` to suit your needs. Remember to inline CSS. You can read more here:

https://github.com/leemunroe/responsive-html-email-template

### Copyright

Do whatever you want

`email.html` is borrowed from
https://github.com/leemunroe/responsive-html-email-template