package mailUtils

import (
	"squad-maker/utils/env"
	"strings"

	mail "github.com/xhit/go-simple-mail/v2"
)

var (
	MailFrom string
)

func init() {
	email, _ := env.GetStr("MAIL_EMAIL")
	if email == "" {
		email, _ = env.GetStr("MAIL_USERNAME")
	}
	name, _ := env.GetStr("MAIL_NAME")

	MailFrom = name + " <" + email + ">"
}

func PrepareNewMail(toName, toEmail, subject, body string, contentType mail.ContentType) *mail.Email {
	return mail.
		NewMSG().
		SetFrom(MailFrom).
		AddTo(toName+" <"+strings.TrimSpace(toEmail)+">").
		SetSubject(subject).
		SetBody(contentType, body)
}
