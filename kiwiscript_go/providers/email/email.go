package email

import "net/smtp"

type Mail struct {
	auth           smtp.Auth
	port           string
	host           string
	address        string
	frontendDomain string
}

func NewMail(username, password, port, host, name, frontendDomain string) *Mail {
	return &Mail{
		auth:           smtp.PlainAuth("", username, password, host),
		port:           port,
		host:           host,
		address:        name + " <" + username + ">",
		frontendDomain: frontendDomain,
	}
}

func (m *Mail) sendMail(to string, subject, body string) error {
	return smtp.SendMail(m.host+":"+m.port, m.auth, m.address, []string{to}, []byte("Subject: "+subject+"\r\n\r\n"+body))
}

func (m *Mail) buildUrl(path, token string) string {
	return "https://" + m.frontendDomain + "/" + path + "/" + token
}
