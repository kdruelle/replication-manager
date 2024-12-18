package mailer

import (
	"crypto/tls"
	"net/smtp"
	"strings"

	"github.com/jordan-wright/email"
)

type Mailer struct {
	Address string
	Auth    smtp.Auth
	TLS     *tls.Config
}

func (m *Mailer) SetAddress(host string) {
	m.Address = host
}

func (m *Mailer) SetSmtpAuth(identity, username, password, host string) {
	m.Auth = smtp.PlainAuth(identity, username, password, host)
}

func (m *Mailer) SetTlsConfig(conf *tls.Config) {
	m.TLS = conf
}

func (m *Mailer) Send(e *email.Email) error {
	return e.Send(m.Address, m.Auth)
}

func (m *Mailer) SendWithTLS(e *email.Email) error {
	return e.SendWithTLS(m.Address, m.Auth, m.TLS)
}

func (m *Mailer) SendEmailMessage(msg string, subj string, From, To, CC string, useTLS bool) error {

	e := email.NewEmail()
	e.From = From
	e.To = strings.Split(To, ",")
	e.Subject = subj
	e.Text = []byte(msg)

	var err error
	if useTLS {
		err = m.SendWithTLS(e)
	} else {
		err = m.Send(e)
	}

	return err
}
