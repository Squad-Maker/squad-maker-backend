package mailUtils

import (
	"squad-maker/utils/env"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

func GetNewSmtpClient() (*mail.SMTPClient, error) {
	host, err := env.GetStr("MAIL_HOST")
	if err != nil {
		return nil, err
	}

	port, _ := env.GetInt("MAIL_PORT")
	username, _ := env.GetStr("MAIL_USERNAME")
	password, _ := env.GetStr("MAIL_PASSWORD")

	smtpClient := mail.NewSMTPClient()
	smtpClient.Host = host
	smtpClient.Port = port
	smtpClient.Username = username
	smtpClient.Password = password

	switch port {
	case 465:
		smtpClient.Encryption = mail.EncryptionSSL
	case 587:
		smtpClient.Encryption = mail.EncryptionSTARTTLS
	default:
		smtpClient.Encryption = mail.EncryptionNone
	}

	smtpClient.Authentication = mail.AuthLogin
	smtpClient.KeepAlive = false
	smtpClient.ConnectTimeout = 30 * time.Second
	smtpClient.SendTimeout = 30 * time.Second

	return smtpClient.Connect()
}
